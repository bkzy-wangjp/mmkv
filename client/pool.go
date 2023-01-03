package client

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	VERSION      = "v0.0.1"    //版本号
	_ReedBufSize = 1024 * 1024 //读信息通道缓存字节数
	//功能码定义(Function Code)
	_FC_Ping            = 0x01 //测试,获取当前时间的UNIX毫秒值
	_FC_Login           = 0x02 //登录
	_FC_WriteSingleKey  = 0x03 //写单个标签
	_FC_WriteMultiKey   = 0x04 //写多个标签
	_FC_ReadSingleKey   = 0x05 //读取单个标签
	_FC_ReadMultiKey    = 0x06 //读取多个标签
	_FC_DeleteSingleKey = 0x07 //删除单个标签
	_FC_DeleteMultiKey  = 0x08 //删除多个标签
	_FC_SelfIncrease    = 0x09 //标签值自增
	_FC_SelfDecrease    = 0x0A //标签自减
	_FC_PipePush        = 0x0B //往管道中压入数据
	_FC_PipeFiFoPull    = 0x0C //先进先出拉取数据
	_FC_PipeFiLoPull    = 0x0D //先进后出拉取数据
	_FC_PipeLenght      = 0x0E //获取管道的长度
	_FC_PipeAll         = 0x0F //一次性读取管道中的所有数据
	_FC_GetKeys         = 0x10 //获取所有已经存在的键
	_FC_GetUsers        = 0x11 //获取所有已经存在的用户名
)

//响应数据的结构
type RespMsg struct {
	Ok   bool        `json:"ok"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

//用户数据的结构
type UserMsg struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

//连接工作池结构
type ConnPool struct {
	WorkerChan chan int         //工作通道
	ReqChan    chan int         //请求队列
	ConnChan   chan *ConnHandel //连接句柄队列
	Conns      []*ConnHandel    //连接池
	size       int              //工作池大小
	maxSec     int64            //最长租借时间(秒)
	address    string           //主机IP
	username   string           //用户名
	password   string           //密码
	timeout    int64            //超时时间,0无限制,单位毫秒
	maxId      int64            //连接池中的连接句柄的最大ID
	Version    string           //版本号
}

/*******************************************************************************
  -功能:实例化工作池
  -参数:
	[hostname]  字符串，输入，服务器的网络地址或机器名,默认值 127.0.0.1
	[username]	用户名,字符串,缺省值 admin
	[password]	密码,字符串,缺省值 admin123
	[langtype]	语言类型,字符串,缺省值 zh-CN
	[port]      端口号,整型,缺省值 9646
	[size]      连接池大小,最小1,最大50
	[timeout]   超时时间(毫秒),默认值 3000
	[max_sec]   最大租借时间(秒),默认值 3600
- 输出:[*ConnPool] 连接池指针
	[error] 创建句柄连接池时候的错误信息
- 备注:
- 时间: 2020年6月27日
*******************************************************************************/
func NewConnPool(params map[string]interface{}) *ConnPool {
	var hostname, username, password, langtype string = "127.0.0.1", "admin", "admin123", _Langtype
	var port, size int = 9646, 50
	var timeout, max_sec int64 = 3000, 60
	for k, v := range params {
		switch k {
		case "hostname":
			hostname, _ = v.(string)
		case "username":
			username, _ = v.(string)
		case "password":
			password, _ = v.(string)
		case "langtype":
			langtype, _ = v.(string)
		case "port":
			port, _ = v.(int)
		case "size":
			size, _ = v.(int)
		case "timeout":
			timeout, _ = v.(int64)
		case "max_sec":
			max_sec, _ = v.(int64)
		}
	}
	if langtype != _Langtype {
		_Langtype = langtype
	}

	if port == 0 {
		port = 9646
	}
	if size < 1 {
		size = 1
	}
	if size > 50 {
		size = 50
	}
	if max_sec < 60 {
		max_sec = 60
	}
	if max_sec <= 0 {
		max_sec = 60
	}
	if len(password) != 32 { //密码必须是MD5后的值
		sum := md5.Sum([]byte(password))
		password = hex.EncodeToString(sum[:])
	}
	address := fmt.Sprintf("%s:%d", hostname, port)

	pool := &ConnPool{
		WorkerChan: make(chan int, size),         //工作者通道
		ReqChan:    make(chan int, 10000),        //请求通道数量
		ConnChan:   make(chan *ConnHandel, size), //池句柄池
		timeout:    timeout,                      //超时时间,毫秒
		size:       size,                         //工作池大小
		maxSec:     max_sec,                      //超时时间
		address:    address,                      //数据库主机名
		username:   username,                     //用户名
		password:   password,                     //密码
		Version:    VERSION,                      //程序版本
	}
	for i := 0; i < size; i++ { //创建句柄连接池
		pool.addConn()
	}
	go pool.run()
	return pool
}

//添加连接
func (p *ConnPool) addConn() error {
	handle, err := newConneHandle(p.timeout, p.username, p.password, p.address)
	p.maxId += 1
	handle.Id = p.maxId
	p.Conns = append(p.Conns, handle) //放入连接池
	p.ConnChan <- handle              //将句柄压入连接池管道
	return err
}

/*******************************************************************************
	- 功能:连接池运行
	- 参数:无
	- 输出:无
	- 备注: 自动释放超时的链接
	- 时间: 2021年4月14日
*******************************************************************************/
func (p *ConnPool) run() {
	defer p.close()
	for {
		for _, c := range p.Conns { //遍历连接池
			if c.ConnServer && !c.Working { //如果处于连接状态
				c.ping() //检查连接是否可用
			}
			if (c.Working && time.Since(c.WorkeAt).Seconds() >= float64(p.maxSec)) || !c.ConnServer { //租用超时或者失去与服务器的连接
				//if c.ConnServer {
				c.Conn.Close() //关闭接口
				//}
				if err := c.dial(p.address); err == nil { //重连接
					c.login(p.username, p.password)
				}
				c.Working = false //停止工作
			}
		}
		time.Sleep(time.Second * 10)
		if len(p.Conns) < p.size { //检查工作池是否还有足够的大小
			for i := 0; i < p.size-len(p.Conns); i++ {
				p.addConn()
			}
		}
	}
}

//关闭连接池
func (p *ConnPool) close() {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	for _, h := range p.Conns {
		h.Conn.Close()
	}
	close(p.ConnChan)
	close(p.ReqChan)
	close(p.WorkerChan)
}

//自由请求接口
//  请求参数:
//   funcCode byte : 功能码
//   msg interface{} : 符合json结构和功能码要求的信息
//反回参数：
//   interface{} : 服务器返回的数据
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func (p *ConnPool) Request(funcCode byte, msg interface{}) (*RespMsg, float64, error) {
	p.ReqChan <- 1         //请求列队加1,如无足够资源，则挂起
	p.WorkerChan <- 1      //工作队列加1,如果无足够资源,则挂起
	handle := <-p.ConnChan //从连接池队列中取出一个连接用于工作
	handle.Working = true
	handle.WorkeAt = time.Now()
	<-p.ReqChan //移除一个请求
	defer func() {
		handle.Working = false
		p.ConnChan <- handle //将连接句柄归还到连接池管道
		<-p.WorkerChan       //工作队列减去1
	}()
	////////////////////////////////////////////////////////////////
	return handle.request(funcCode, msg)
}

//获取所有的非系统KEYS
//  请求参数:
//   无
//反回参数：
//  interface{} : 标签所对应的数据,单个标签时为data，多个标签时为[]{Ok:bool,Msg:string,Data:interface{}}
//  float64 : 执行耗时,单位秒
//  error : 错误信息
func (p *ConnPool) GetKeys() ([]interface{}, float64, error) {
	p.ReqChan <- 1         //请求列队加1,如无足够资源，则挂起
	p.WorkerChan <- 1      //工作队列加1,如果无足够资源,则挂起
	handle := <-p.ConnChan //从连接池队列中取出一个连接用于工作
	handle.Working = true
	handle.WorkeAt = time.Now()
	<-p.ReqChan //移除一个请求
	defer func() {
		handle.Working = false
		p.ConnChan <- handle //将连接句柄归还到连接池管道
		<-p.WorkerChan       //工作队列减去1
	}()
	////////////////////////////////////////////////////////////////
	var data []byte
	msg, sec, err := handle.request(_FC_GetKeys, data)
	if err != nil {
		return nil, sec, err
	} else {
		//fmt.Printf("原始数据:%+v\n", msg.Data)
		keysmp, _ := msg.Data.([]interface{})
		return keysmp, sec, err
	}
}

//获取所有的用户名列表
//  请求参数:
//   无
//反回参数：
//   interface{} : 标签所对应的数据,单个标签时为data，多个标签时为[]{Ok:bool,Msg:string,Data:interface{}}
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func (p *ConnPool) GetUsers() ([]string, float64, error) {
	p.ReqChan <- 1         //请求列队加1,如无足够资源，则挂起
	p.WorkerChan <- 1      //工作队列加1,如果无足够资源,则挂起
	handle := <-p.ConnChan //从连接池队列中取出一个连接用于工作
	handle.Working = true
	handle.WorkeAt = time.Now()
	<-p.ReqChan //移除一个请求
	defer func() {
		handle.Working = false
		p.ConnChan <- handle //将连接句柄归还到连接池管道
		<-p.WorkerChan       //工作队列减去1
	}()
	////////////////////////////////////////////////////////////////
	var data []byte
	msg, sec, err := handle.request(_FC_GetUsers, data)
	if err != nil {
		return nil, sec, err
	} else {
		//fmt.Printf("原始数据:%+v\n", msg.Data)
		var users []string
		usermp, _ := msg.Data.([]interface{})
		for _, v := range usermp {
			users = append(users, v.(string))
		}
		return users, sec, err
	}
}

//读取标签
//  请求参数:
//   tags ...string : 标签名
//反回参数：
//   interface{} : 标签所对应的数据,单个标签时为data，多个标签时为[]{Ok:bool,Msg:string,Data:interface{}}
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func (p *ConnPool) Read(tags ...string) (interface{}, float64, error) {
	p.ReqChan <- 1         //请求列队加1,如无足够资源，则挂起
	p.WorkerChan <- 1      //工作队列加1,如果无足够资源,则挂起
	handle := <-p.ConnChan //从连接池队列中取出一个连接用于工作
	handle.Working = true
	handle.WorkeAt = time.Now()
	<-p.ReqChan //移除一个请求
	defer func() {
		handle.Working = false
		p.ConnChan <- handle //将连接句柄归还到连接池管道
		<-p.WorkerChan       //工作队列减去1
	}()
	////////////////////////////////////////////////////////////////
	if len(tags) == 1 {
		msg, sec, err := handle.request(_FC_ReadSingleKey, tags[0])
		if err != nil {
			return nil, sec, err
		} else {
			return msg.Data, sec, err
		}
	} else {
		msg, sec, err := handle.request(_FC_ReadMultiKey, tags)
		if err != nil {
			return nil, sec, err
		} else {
			return msg.Data, sec, err
		}
	}
}

//写标签
//  请求参数:
//   kvs map[string]interface{} : 需要写入的数据
//反回参数：
//   interface{}: 返回的数据
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func (p *ConnPool) Write(key string, val interface{}) (interface{}, float64, error) {
	p.ReqChan <- 1         //请求列队加1,如无足够资源，则挂起
	p.WorkerChan <- 1      //工作队列加1,如果无足够资源,则挂起
	handle := <-p.ConnChan //从连接池队列中取出一个连接用于工作
	handle.Working = true
	handle.WorkeAt = time.Now()
	<-p.ReqChan //移除一个请求
	defer func() {
		handle.Working = false
		p.ConnChan <- handle //将连接句柄归还到连接池管道
		<-p.WorkerChan       //工作队列减去1
	}()
	////////////////////////////////////////////////////////////////
	msg, sec, err := handle.request(_FC_WriteSingleKey, map[string]interface{}{key: val})
	if err != nil {
		return nil, sec, err
	} else {
		return msg.Data, sec, err
	}
}

//写多个标签
//  请求参数:
//   kvs map[string]interface{} : 需要写入的数据
//反回参数：
//   interface{}: 返回的数据
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func (p *ConnPool) Writes(kvs map[string]interface{}) (interface{}, float64, error) {
	p.ReqChan <- 1         //请求列队加1,如无足够资源，则挂起
	p.WorkerChan <- 1      //工作队列加1,如果无足够资源,则挂起
	handle := <-p.ConnChan //从连接池队列中取出一个连接用于工作
	handle.Working = true
	handle.WorkeAt = time.Now()
	<-p.ReqChan //移除一个请求
	defer func() {
		handle.Working = false
		p.ConnChan <- handle //将连接句柄归还到连接池管道
		<-p.WorkerChan       //工作队列减去1
	}()
	////////////////////////////////////////////////////////////////
	msg, sec, err := handle.request(_FC_WriteMultiKey, kvs)
	if err != nil {
		return nil, sec, err
	} else {
		return msg.Data, sec, err
	}
}

//删除标签
//  请求参数:
//   tags ...string : 标签名
//反回参数：
//   bool : 删除的结果
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func (p *ConnPool) Delete(tags ...string) (bool, float64, error) {
	p.ReqChan <- 1         //请求列队加1,如无足够资源，则挂起
	p.WorkerChan <- 1      //工作队列加1,如果无足够资源,则挂起
	handle := <-p.ConnChan //从连接池队列中取出一个连接用于工作
	handle.Working = true
	handle.WorkeAt = time.Now()
	<-p.ReqChan //移除一个请求
	defer func() {
		handle.Working = false
		p.ConnChan <- handle //将连接句柄归还到连接池管道
		<-p.WorkerChan       //工作队列减去1
	}()
	////////////////////////////////////////////////////////////////
	_, sec, err := handle.request(_FC_DeleteMultiKey, tags)
	if err != nil {
		return false, sec, err
	} else {
		return true, sec, err
	}
}

//服务器时间
//  请求参数:无
//反回参数：
//   time.Time : 时间
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func (p *ConnPool) ServerTime() (time.Time, float64, error) {
	p.ReqChan <- 1         //请求列队加1,如无足够资源，则挂起
	p.WorkerChan <- 1      //工作队列加1,如果无足够资源,则挂起
	handle := <-p.ConnChan //从连接池队列中取出一个连接用于工作
	handle.Working = true
	handle.WorkeAt = time.Now()
	<-p.ReqChan //移除一个请求
	defer func() {
		handle.Working = false
		p.ConnChan <- handle //将连接句柄归还到连接池管道
		<-p.WorkerChan       //工作队列减去1
	}()
	////////////////////////////////
	var data []byte
	msg, sec, err := handle.request(_FC_Ping, data)

	if err != nil {
		return time.Now(), sec, err
	} else {
		//fmt.Printf("获取到的时间:%s\n", tstr)
		fmicroSec, ok := msg.Data.(float64)
		if ok {
			microSec := int64(fmicroSec)
			t := time.Unix(microSec/1e6, microSec%1e6*1e3)
			return t, sec, err
		} else {
			return time.Now(), sec, err
		}
	}
}

//标签自增
//  请求参数:
//   tag string : 标签名
//反回参数：
//   int64 : 当前值
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func (p *ConnPool) SelfIncrease(tag string) (int64, float64, error) {
	p.ReqChan <- 1         //请求列队加1,如无足够资源，则挂起
	p.WorkerChan <- 1      //工作队列加1,如果无足够资源,则挂起
	handle := <-p.ConnChan //从连接池队列中取出一个连接用于工作
	handle.Working = true
	handle.WorkeAt = time.Now()
	<-p.ReqChan //移除一个请求
	defer func() {
		handle.Working = false
		p.ConnChan <- handle //将连接句柄归还到连接池管道
		<-p.WorkerChan       //工作队列减去1
	}()
	////////////////////////////////
	msg, sec, err := handle.request(_FC_SelfIncrease, tag)
	if err != nil {
		return 0, sec, err
	} else {
		val, ok := msg.Data.(int64)
		if !ok {
			fval, ok := msg.Data.(float64)
			if !ok {
				return 0, sec, fmt.Errorf(i18n("data_type_fail_add"), reflect.TypeOf(msg.Data))
			}
			val = int64(fval)
		}
		return val, sec, err
	}
}

//标签自减
//  请求参数:
//   tag string : 标签名
//反回参数：
//   int64 : 当前值
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func (p *ConnPool) SelfDecrease(tag string) (int64, float64, error) {
	p.ReqChan <- 1         //请求列队加1,如无足够资源，则挂起
	p.WorkerChan <- 1      //工作队列加1,如果无足够资源,则挂起
	handle := <-p.ConnChan //从连接池队列中取出一个连接用于工作
	handle.Working = true
	handle.WorkeAt = time.Now()
	<-p.ReqChan //移除一个请求
	defer func() {
		handle.Working = false
		p.ConnChan <- handle //将连接句柄归还到连接池管道
		<-p.WorkerChan       //工作队列减去1
	}()
	////////////////////////////////
	msg, sec, err := handle.request(_FC_SelfDecrease, tag)
	if err != nil {
		return 0, sec, err
	} else {
		val, ok := msg.Data.(int64)
		if !ok {
			fval, ok := msg.Data.(float64)
			if !ok {
				return 0, sec, fmt.Errorf(i18n("data_type_fail_add"), reflect.TypeOf(msg.Data))
			}
			val = int64(fval)
		}
		return val, sec, err
	}
}

//压入管道
//  请求参数:
//   key string 标签名
//   value interface{} 数值
//反回参数：
//   int64: 当前管道长度
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func (p *ConnPool) PipePush(key string, value interface{}) (int64, float64, error) {
	p.ReqChan <- 1         //请求列队加1,如无足够资源，则挂起
	p.WorkerChan <- 1      //工作队列加1,如果无足够资源,则挂起
	handle := <-p.ConnChan //从连接池队列中取出一个连接用于工作
	handle.Working = true
	handle.WorkeAt = time.Now()
	<-p.ReqChan //移除一个请求
	defer func() {
		handle.Working = false
		p.ConnChan <- handle //将连接句柄归还到连接池管道
		<-p.WorkerChan       //工作队列减去1
	}()
	////////////////////////////////////////////////////////////////
	msg, sec, err := handle.request(_FC_PipePush, map[string]interface{}{key: value})
	if err != nil {
		return 0, sec, err
	} else {
		val, ok := msg.Data.(int64)
		if !ok {
			fval, ok := msg.Data.(float64)
			if !ok {
				return 0, sec, fmt.Errorf(i18n("data_assert_fail"), reflect.TypeOf(msg.Data))
			}
			val = int64(fval)
		}
		return val, sec, err
	}
}

//获取管道当前的长度
//  请求参数:
//   tag string : 标签名
//反回参数：
//   int64 : 当前管道长度
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func (p *ConnPool) PipeLength(tag string) (int64, float64, error) {
	p.ReqChan <- 1         //请求列队加1,如无足够资源，则挂起
	p.WorkerChan <- 1      //工作队列加1,如果无足够资源,则挂起
	handle := <-p.ConnChan //从连接池队列中取出一个连接用于工作
	handle.Working = true
	handle.WorkeAt = time.Now()
	<-p.ReqChan //移除一个请求
	defer func() {
		handle.Working = false
		p.ConnChan <- handle //将连接句柄归还到连接池管道
		<-p.WorkerChan       //工作队列减去1
	}()
	////////////////////////////////
	msg, sec, err := handle.request(_FC_PipeLenght, tag)
	if err != nil {
		return 0, sec, err
	} else {
		val, ok := msg.Data.(int64)
		if !ok {
			fval, ok := msg.Data.(float64)
			if !ok {
				return 0, sec, fmt.Errorf(i18n("data_assert_fail"), reflect.TypeOf(msg.Data))
			}
			val = int64(fval)
		}
		return val, sec, err
	}
}

//从管道中拉取数据
//  请求参数:
//   tag string : 标签名
//   ptype string : 拉取方式, "FIFO"、"LILO"或"FILO"、"LIFO"(不区分大小写)
//反回参数:
//   int64 : 剩余长度
//   interface{} : 获取到的数据
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func (p *ConnPool) PipePull(tag, ptype string) (int64, interface{}, float64, error) {
	p.ReqChan <- 1         //请求列队加1,如无足够资源，则挂起
	p.WorkerChan <- 1      //工作队列加1,如果无足够资源,则挂起
	handle := <-p.ConnChan //从连接池队列中取出一个连接用于工作
	handle.Working = true
	handle.WorkeAt = time.Now()
	<-p.ReqChan //移除一个请求
	defer func() {
		handle.Working = false
		p.ConnChan <- handle //将连接句柄归还到连接池管道
		<-p.WorkerChan       //工作队列减去1
	}()
	////////////////////////////////
	var fc byte
	switch strings.ToLower(ptype) {
	case "fifo", "lilo":
		fc = _FC_PipeFiFoPull
	case "filo", "lifo":
		fc = _FC_PipeFiLoPull
	case "all":
		fc = _FC_PipeAll
	default:
		return 0, 0, 0., fmt.Errorf(i18n("log_pipe_pull_type_err"), ptype)
	}
	msg, sec, err := handle.request(fc, tag)
	if err != nil {
		return 0, 0, sec, err
	} else {
		length, _ := strconv.ParseInt(msg.Msg, 10, 64)
		return length, msg.Data, sec, err
	}
}

//从管道中拉取所有数据
//请求参数:
//   tag string : 标签名
//反回参数:
//   interface{} : 获取到的数据
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func (p *ConnPool) PipeAll(tag string) (interface{}, float64, error) {
	p.ReqChan <- 1         //请求列队加1,如无足够资源，则挂起
	p.WorkerChan <- 1      //工作队列加1,如果无足够资源,则挂起
	handle := <-p.ConnChan //从连接池队列中取出一个连接用于工作
	handle.Working = true
	handle.WorkeAt = time.Now()
	<-p.ReqChan //移除一个请求
	defer func() {
		handle.Working = false
		p.ConnChan <- handle //将连接句柄归还到连接池管道
		<-p.WorkerChan       //工作队列减去1
	}()
	////////////////////////////////
	msg, sec, err := handle.request(_FC_PipeAll, tag)
	if err != nil {
		return nil, sec, err
	} else {
		return msg.Data, sec, err
	}
}
