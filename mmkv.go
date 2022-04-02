package mmkv

import (
	"encoding/json"
	"fmt"
	"net"
	"reflect"

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
//    }
func Run(users map[string]string, cfg map[string]interface{}) error {
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
func i18n(ecode string, lang ...string) string {
	if len(lang) == 0 {
		return Db.getErrorMsg(ecode)
	} else {
		ecode, ok := textDictionary[ecode]
		if !ok {
			ecode = unDefined
		}
		msg, ok := ecode[lang[0]]
		if !ok {
			msg = fmt.Sprintf("Undefined languige type:%s", lang[0])
		}
		return msg
	}
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
