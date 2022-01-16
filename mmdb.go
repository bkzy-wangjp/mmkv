package mmdb

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	crc16 "github.com/bkzy-wangjp/CRC16"
)

const (
	VERSION         = "v0.0.1"    //版本号
	_ReedBufSize    = 1024 * 1024 //读信息通道缓存字节数
	_UserPrefix     = "users."    //用户名前缀
	_TextDictPrefix = "text."     //文本字典前缀
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

//通讯句柄结构
type ConnHandel struct {
	Id        int64     //序号
	Conn      net.Conn  //通讯连接
	Logged    bool      //是否已登录标志
	TxBytes   int64     //发送字节数
	RxBytes   int64     //接收字节数
	TxTimes   int64     //发送次数
	RxTimes   int64     //接收次数
	User      string    //用户信息
	Closed    bool      //连接已关闭
	LogAt     time.Time //登录时间
	CreatedAt time.Time //创建时间
	CloseAt   time.Time //关闭时间
}

//运行数据库结构
type MemoryDb struct {
	sync.Map
	Langtype string //language type,like:en-US,zh-CN
}

var (
	MmDb     MemoryDb      //内存数据库
	ConnPool []*ConnHandel //连接池
)

//运行数据库系统
func Run(users map[string]string, port int64, ips ...string) error {
	MmDb.MmAddUsers(users) //初始化用户表
	MmDb.Langtype = "en-US"
	ip := ""
	if len(ips) > 0 {
		ip = ips[0]
	}
	if port == 0 {
		port = 9646
	}
	address := fmt.Sprintf("%s:%d", ip, port)
	l, err := net.Listen("tcp", address)
	if err != nil {
		logs.Error(i18n("log_err_listen_faile"), err.Error())
		return err
	}
	logs.Info(i18n("log_info_listen_port"), address)
	var id int64 = 0
	for {
		id++
		conn, err := l.Accept()
		if err != nil {
			logs.Error(i18n("log_err_new_conn"), err.Error())
			continue
		}
		logs.Info(i18n("log_info_new_conn"), conn.RemoteAddr())
		handle := newConneHandle(conn)
		handle.Id = id
		ConnPool = append(ConnPool, handle)
		go handle.handleConn()
		//printConnPool()
	}
}

func printConnPool() {
	for i, handle := range ConnPool {
		logs.Debug("序号%d,信息:%+v", i, handle)
	}
}

//分割已经校验通过的数据
func splitData(hex_data []byte) (funcCode byte, reserve byte, data []byte, err error) {
	if len(hex_data) < 2 {
		err = fmt.Errorf(i18n("log_err_len_less_2byte"))
		return
	}
	funcCode = hex_data[0]
	reserve = hex_data[1]
	data = hex_data[2:]
	return
}

//编码响应字符串
func encodeResponse(funcCode, reverse byte, data []byte) []byte {
	var resp []byte
	resp = append(resp, funcCode, reverse)
	resp = append(resp, data...)
	return crc16.BytesAndCrcSum(resp, true)
}

//解析单个key
func decodeKey(hex_data []byte) (key string, err error) {
	err = json.Unmarshal(hex_data, &key)
	return
}

//解析多个key
func decodeKeys(hex_data []byte) (keys []string, err error) {
	err = json.Unmarshal(hex_data, &keys)
	return
}

//解析key\value Map
func decodeMap(hex_data []byte) (kvs map[string]interface{}, err error) {
	err = json.Unmarshal(hex_data, &kvs)
	return
}

//解析用户信息
func decodeUserMsg(hex_data []byte) (user UserMsg, err error) {
	err = json.Unmarshal(hex_data, &user)
	return
}

//通过错误码获取错误信息
func i18n(ecode string) string {
	return MmDb.I18n(ecode)
}
