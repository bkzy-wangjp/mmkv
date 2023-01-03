package client

import "time"

var (
	_POOL = new(ConnPool)
)

/*
新建内存数据库的连接池:
{
	"hostname":"字符串，输入，服务器的网络地址或机器名,默认值 127.0.0.1"
	"username":"用户名,字符串,缺省值 admin"
	"password":"密码,字符串,缺省值 admin123"
	"langtype":"语言类型,字符串,缺省值 zh-CN"
	"port"    :"端口号,整型,缺省值 9646"
	"size"    :"连接池大小,最小1,最大50"
	"timeout" :"超时时间(毫秒),默认值 3000"
	"max_sec" :"最大租借时间(秒),默认值 3600"
}
*/
func NewClient(params map[string]interface{}) {
	_POOL = NewConnPool(params)
}

//自由请求接口
//  请求参数:
//    funcCode byte : 功能码
//    msg interface{} : 符合json结构和功能码要求的信息
//反回参数：
//  interface{} : 服务器返回的数据
//  float64 : 执行耗时,单位秒
//  error : 错误信息
func Request(funcCode byte, msg interface{}) (*RespMsg, float64, error) {
	return _POOL.Request(funcCode, msg)
}

//获取所有的非系统KEYS
//  请求参数:
//   无
//反回参数：
//   interface{} : 标签所对应的数据,单个标签时为data，多个标签时为[]{Ok:bool,Msg:string,Data:interface{}}
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func GetKeys() ([]interface{}, float64, error) {
	return _POOL.GetKeys()
}

//获取所有的用户名列表
//  请求参数:
//    无
//反回参数：
//    interface{} : 标签所对应的数据,单个标签时为data，多个标签时为[]{Ok:bool,Msg:string,Data:interface{}}
//    float64 : 执行耗时,单位秒
//    error : 错误信息
func GetUsers() ([]string, float64, error) {
	return _POOL.GetUsers()
}

//读取标签
//  请求参数:
//   tags ...string : 标签名
//反回参数：
//   interface{} : 标签所对应的数据,单个标签时为data，多个标签时为[]{Ok:bool,Msg:string,Data:interface{}}
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func Read(tags ...string) (interface{}, float64, error) {
	return _POOL.Read(tags...)
}

//写单个标签
//  请求参数:
//   kvs map[string]interface{} : 需要写入的数据
//反回参数：
//   interface{}: 返回的数据
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func Write(key string, value interface{}) (interface{}, float64, error) {
	return _POOL.Write(key, value)
}

//写多个标签
//  请求参数:
//   kvs map[string]interface{} : 需要写入的数据
//反回参数：
//   interface{}: 返回的数据
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func Writes(kvs map[string]interface{}) (interface{}, float64, error) {
	return _POOL.Writes(kvs)
}

//删除标签
//  请求参数:
//   tags ...string : 标签名
//反回参数：
//   bool : 删除的结果
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func Delete(tags ...string) (bool, float64, error) {
	return _POOL.Delete(tags...)
}

//服务器时间
//  请求参数:无
//反回参数：
//   time.Time : 时间
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func ServerTime() (time.Time, float64, error) {
	return _POOL.ServerTime()
}

//标签自增
//  请求参数:
//   tag string : 标签名
//反回参数：
//   int64 : 当前值
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func SelfIncrease(tag string) (int64, float64, error) {
	return _POOL.SelfIncrease(tag)
}

//标签自减
//  请求参数:
//   tag string : 标签名
//反回参数：
//   int64 : 当前值
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func SelfDecrease(tag string) (int64, float64, error) {
	return _POOL.SelfDecrease(tag)
}

//压入管道
//  请求参数:
//   key string 标签名
//   value interface{} 数值
//反回参数：
//   int64: 当前管道长度
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func PipePush(key string, value interface{}) (int64, float64, error) {
	return _POOL.PipePush(key, value)
}

//获取管道当前的长度
//  请求参数:
//   tag string : 标签名
//反回参数：
//   int64 : 当前管道长度
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func PipeLength(tag string) (int64, float64, error) {
	return _POOL.PipeLength(tag)
}

//从管道中拉取数据
//  请求参数:
//   tag string : 标签名
//   ptype string : 拉取方式, "FIFO"、"LILO"或"FILO"、"LIFO"(不区分大小写)
//反回参数:
//   int64 : 剩余长度
//   interface{} : 获取到的数据
//   float64 : 执行耗时,单位秒
// error : 错误信息
func PipePull(tag, ptype string) (int64, interface{}, float64, error) {
	return _POOL.PipePull(tag, ptype)
}

//从管道中拉取所有数据
//  请求参数:
//   tag string : 标签名
//反回参数:
//   interface{} : 获取到的数据
//   float64 : 执行耗时,单位秒
//   error : 错误信息
func PipeAll(tag string) (interface{}, float64, error) {
	return _POOL.PipeAll(tag)
}
