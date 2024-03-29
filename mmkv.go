package mmkv

import (
	"encoding/json"
	"fmt"
	"net"
	"path"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	crc16 "github.com/bkzy-wangjp/CRC16"
)

//运行数据库系统
//输入:
//    users map[string]string:{username:password}
//    cfg map[string]string:{
//        "port":9646 //端口号
//        "ip": "127.0.0.1"//IP地址
//        "languige":"zh-CN"//语言类型
//        "loglevel":8
//		  "logpath":"../log"
// 		  "logsize":100000
//        "logsavedays":180
//        "logshowpath":false
//    }
func Run(users map[string]string, cfg map[string]interface{}) error {
	logpath, ok := cfg["logpath"]
	if !ok {
		logpath = "./log/mmkv.log"
	} else {
		logpath = path.Join(fmt.Sprint(logpath), "mmkv.log")
	}
	logsize, ok := cfg["logsize"]
	if !ok {
		logsize = 100000
	}
	logsavedays, ok := cfg["logsavedays"]
	if !ok {
		logsavedays = 180
	}
	loglevel, ok := cfg["loglevel"]
	if !ok {
		loglevel = 8
	}
	logshowpath, ok := cfg["logshowpath"]
	if !ok {
		logshowpath = false
	}
	showpath, ok := logshowpath.(bool)
	if !ok {
		showpath = false
	}
	logs.SetLogger(logs.AdapterConsole, fmt.Sprintf(`{"level":%d,"color":true}`, loglevel)) //屏幕输出设置
	logset := fmt.Sprintf(`{"filename":"%s","level":%d,"maxlines":%d,"maxsize":0,"daily":true,"maxdays":%d}`, logpath, loglevel, logsize, logsavedays)
	logs.SetLogger(logs.AdapterMultiFile, logset) //文件输出设置
	logs.EnableFuncCallDepth(showpath)

	logs.Info(i18n("内存数据库启动"))
	logs.Debug("%s:%+v", i18n("启动参数"), cfg)

	languige, ok := cfg["languige"]
	if !ok {
		languige = "zh-CN"
	}
	Db.init(users, languige.(string))
	ip, ok := cfg["ip"]
	if !ok {
		ip = "127.0.0.1"
	}
	port, ok := cfg["port"]
	if !ok {
		port = 9646
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

	go saveMemStats() //记录内存使用情况

	var id int64 = 0
	for {
		id++
		conn, err := l.Accept()
		if err != nil {
			logs.Error(i18n("log_err_new_conn"), err.Error())
			continue
		}
		//logs.Info(i18n("log_info_new_conn"), conn.RemoteAddr())
		handle := newConneHandle(conn)
		handle.Id = id
		ConnPool = append(ConnPool, handle)
		go handle.handleConn()
		//printConnPool()
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
	key = strings.ToLower(key)
	return
}

//解析多个key
func decodeKeys(hex_data []byte) (keys []string, err error) {
	err = json.Unmarshal(hex_data, &keys)
	for i, k := range keys {
		keys[i] = strings.ToLower(k)
	}
	return
}

//解析key\value Map
func decodeMap(hex_data []byte) (kvs map[string]interface{}, err error) {
	maps := make(map[string]interface{})
	kvs = make(map[string]interface{})
	err = json.Unmarshal(hex_data, &maps)
	for k, v := range maps {
		kvs[strings.ToLower(k)] = v
	}
	return
}

//解析用户信息
func decodeUserMsg(hex_data []byte) (user UserMsg, err error) {
	err = json.Unmarshal(hex_data, &user)
	return
}

//通过错误码获取错误信息
func i18n(msg string, lang ...string) string {
	langtype := Db.Langtype
	if len(lang) > 0 {
		langtype = lang[0]
	}

	ecode, ok := textDictionary[msg]
	if !ok {
		return msg
	}
	rmsg, ok := ecode[langtype]
	if !ok {
		rmsg = msg
	}
	return rmsg
}

//连接反馈信息格式化
func MakeRespMsg(ok bool, msg string, data interface{}) RespMsg {
	var resp RespMsg
	resp.Ok = ok
	resp.Msg = msg
	resp.Data = data
	return resp
}

//数据自加运算
//如果数据类型不可进行加减运算,返回错误信息
func selfAdd(oldval interface{}, increment int64) (interface{}, error) {
	vtype := reflect.TypeOf(oldval)
	switch vtype.Kind() {
	case reflect.Uint:
		return int64(oldval.(uint)) + increment, nil
	case reflect.Uint8:
		return int64(oldval.(uint8)) + increment, nil
	case reflect.Uint16:
		return int64(oldval.(uint16)) + increment, nil
	case reflect.Uint32:
		return int64(oldval.(uint32)) + increment, nil
	case reflect.Uint64:
		return int64(oldval.(uint64)) + increment, nil
	case reflect.Int:
		return int64(oldval.(int)) + increment, nil
	case reflect.Int8:
		return int64(oldval.(int8)) + increment, nil
	case reflect.Int16:
		return int64(oldval.(int16)) + increment, nil
	case reflect.Int32:
		return int64(oldval.(int32)) + increment, nil
	case reflect.Int64:
		return oldval.(int64) + increment, nil
	case reflect.Float32:
		return oldval.(float32) + float32(increment), nil
	case reflect.Float64:
		return oldval.(float64) + float64(increment), nil
	}
	return nil, fmt.Errorf(i18n("data_type_fail_add"), vtype)
}

//判断是否数字类型的值
func isNumber(value interface{}) (bool, reflect.Type) {
	if value == nil {
		return false, nil
	}
	vtype := reflect.TypeOf(value)
	switch vtype.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true, vtype
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true, vtype
	case reflect.Float32, reflect.Float64:
		return true, vtype
	}
	return false, vtype
}

//判断是否系统保留字
func isSysReservedKey(key string) bool {
	reservedkeys := []string{_KeysDict, _UsersDict}
	for _, k := range reservedkeys {
		if k == key {
			return true
		}
	}
	return false
}

//保存内存信息
func saveMemStats() {
	for {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		Db.MmWriteSingle("sys.memstats.alloc", float64(m.Alloc)/(1024*1024))                  //当前堆上对象占用的内存大小
		Db.MmWriteSingle("sys.memstats.totalalloc", float64(m.TotalAlloc)/(1024*1024))        //堆上总共分配出的内存大小
		Db.MmWriteSingle("sys.memstats.totalfree", float64(m.TotalAlloc-m.Alloc)/(1024*1024)) //堆上空闲的内存大小
		Db.MmWriteSingle("sys.memstats.sys", float64(m.Sys)/(1024*1024))                      //程序从操作系统总共申请的内存大小
		Db.MmWriteSingle("sys.memstats.numgc", m.NumGC)                                       //垃圾回收运行的次数

		time.Sleep(60 * time.Second) //每分钟记录一次
	}
}
