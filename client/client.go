package client

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	VERSION      = "v0.0.1"    //版本号
	_ReedBufSize = 1024 * 1024 //读信息通道缓存字节数
	//功能码定义(Function Code)
	_FC_WriteSingleKey  = 1 //写单个标签
	_FC_WriteMultiKey   = 2 //写多个标签
	_FC_ReadSingleKey   = 3 //读取单个标签
	_FC_ReadMultiKey    = 4 //读取多个标签
	_FC_DeleteSingleKey = 5 //删除单个标签
	_FC_DeleteMultiKey  = 6 //删除多个标签
	_FC_Login           = 7 //登录
	_FC_Ping            = 8 //测试,获取当前时间的UNIX毫秒值
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
- 功能:实例化工作池
- 参数:
	[hostname]  字符串，输入，GOLDEN 数据平台服务器的网络地址或机器名
	[username]	用户名,字符串,缺省值 sa
	[password]	密码,字符串,缺省值 golden
	[port]      端口号,整型,缺省值 6327
	[cap]       连接池大小,最小1,最大50
	[max_sec]   最大租借时间(秒)
- 输出:[*ConnPool] 连接池指针
	[error] 创建句柄连接池时候的错误信息
- 备注:
- 时间: 2020年6月27日
*******************************************************************************/
func NewConnPool(params map[string]interface{}) (*ConnPool, error) {
	var hostname, username, password, langtype string = "127.0.0.1", "admin", "admin123", _Langtype
	var port, size int = 9646, 50
	var timeout, max_sec int64 = 3000, 3600
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
		max_sec = 3600
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
	var err error
	for i := 0; i < size; i++ { //创建句柄连接池
		pool.addConn()
	}
	go pool.run()
	return pool, err
}

//添加连接
func (p *ConnPool) addConn() error {
	handle, err := newConneHandle(p.username, p.password, p.address)
	if err != nil {
		return err
	}
	p.maxId += 1
	handle.Id = p.maxId
	p.Conns = append(p.Conns, handle) //放入连接池
	p.ConnChan <- handle              //将句柄压入连接池管道
	return nil
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
			if (c.Working && time.Since(c.WorkeAt).Seconds() >= float64(p.maxSec)) || !c.ConnServer { //租用超时或者失去与服务器的连接
				c.Working = false //停止工作
				c.Conn.Close()    //关闭接口
				c.dial(p.address) //重连接
			}
		}
		time.Sleep(time.Second * 60)
		if len(p.Conns) < p.size { //检查工作池是否还有足够的大小
			for i := 0; i < p.size-len(p.Conns); i++ {
				p.addConn()
			}
		}
	}
}

//关闭连接池
func (p *ConnPool) close() {
	for _, h := range p.Conns {
		h.Conn.Close()
	}
	close(p.ConnChan)
	close(p.ReqChan)
	close(p.WorkerChan)
}

//自由请求接口
//请求参数:
// funcCode byte : 功能码
// msg interface{} : 符合json结构和功能码要求的信息
//反回参数：
// interface{} : 服务器返回的数据
// float64 : 执行耗时,单位秒
// error : 错误信息
func (p *ConnPool) Request(funcCode byte, msg interface{}) (interface{}, float64, error) {
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
	return handle.request(funcCode, msg)
}

//读取标签
//请求参数:
// tags ...string : 标签名
//反回参数：
// interface{} : 标签所对应的数据,单个标签时为data，多个标签时为[]{Ok:bool,Msg:string,Data:interface{}}
// float64 : 执行耗时,单位秒
// error : 错误信息
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
	if len(tags) == 1 {
		return handle.request(_FC_ReadSingleKey, tags[0])
	} else {
		return handle.request(_FC_ReadMultiKey, tags)
	}
}

//写标签
//请求参数:
// kvs map[string]interface{} : 需要写入的数据
//反回参数：
// bool: 写入的结果
// float64 : 执行耗时,单位秒
// error : 错误信息
func (p *ConnPool) Write(kvs map[string]interface{}) (bool, float64, error) {
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
	msg, sec, err := handle.request(_FC_WriteMultiKey, kvs)
	state, _ := msg.(bool)
	return state, sec, err
}

//删除标签
//请求参数:
// tags ...string : 标签名
//反回参数：
// bool : 删除的结果
// float64 : 执行耗时,单位秒
// error : 错误信息
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
	msg, sec, err := handle.request(_FC_DeleteMultiKey, tags)
	state, _ := msg.(bool)
	return state, sec, err
}

//服务器时间
//请求参数:
// tags ...string : 标签名
//反回参数：
// time.Time : 时间
// float64 : 执行耗时,单位秒
// error : 错误信息
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
	var data []byte
	msg, sec, err := handle.request(_FC_Ping, data)
	microSec, _ := msg.(int64)
	t := time.Unix(microSec/1e6, microSec%1e6*1e3)
	return t, sec, err
}
