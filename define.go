package mmkv

import (
	"net"
	"sync"
	"time"
)

const (
	VERSION      = "v0.0.1"         //版本号
	_ReedBufSize = 1024 * 1024      //读信息通道缓存字节数
	_UsersDict   = "sys.users.dict" //用户字典
	_KeysDict    = "sys.keys.dict"  //键名称字典

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

	_FC_GetKeys  = 0x10 //获取所有已经存在的键
	_FC_GetUsers = 0x11 //获取所有已经存在的用户名
)

var (
	Db       MemoryKeyValueMap //内存数据库
	ConnPool []*ConnHandel     //连接池

	_FC_MAP = map[byte]string{
		0x01: "Ping",
		0x02: "Login",
		0x03: "WriteSingleKey",
		0x04: "WriteMultiKey",
		0x05: "ReadSingleKey",
		0x06: "ReadMultiKey",
		0x07: "DeleteSingleKey",
		0x08: "DeleteMultiKey",
		0x09: "SelfIncrease",
		0x0A: "SelfDecrease",
		0x0B: "PipePush",
		0x0C: "PipeFiFoPull",
		0x0D: "PipeFiLoPull",
		0x0E: "PipeLenght",
		0x10: "GetKeys",
		0x11: "GetUsers",
	}
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

//用户字典
type UserDict struct {
	Password string //Md5后的密码
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
type MemoryKeyValueMap struct {
	sync.Map
	Langtype string //language type,like:en-US,zh-CN
}
