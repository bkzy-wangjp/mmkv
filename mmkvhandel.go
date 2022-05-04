package mmkv

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	crc16 "github.com/bkzy-wangjp/CRC16"
)

//新建通讯句柄结构
func newConneHandle(conn net.Conn) *ConnHandel {
	handle := new(ConnHandel)
	handle.Conn = conn
	handle.CreatedAt = time.Now()
	return handle
}

//连接句柄
func (h *ConnHandel) handleConn() {
	var buf []byte
	read_start := make(chan bool, 10)
	defer func() {
		h.Conn.Close()
		h.Closed = true
		h.CloseAt = time.Now()
		close(read_start)
	}()
	isend := true //是否读取结束
	for {
		//read from the connection
		var rbuf = make([]byte, _ReedBufSize)
		n, err := h.Conn.Read(rbuf) //读取缓冲通道，如果无数据则挂起
		if !isend {
			read_start <- true //开始读标记
		}
		if err != nil {
			if err == io.EOF || strings.Contains(err.Error(), "close") || strings.Contains(err.Error(), "aborted") { //客户端关闭
				logs.Info(i18n("log_info_client_shutdown"), h.Conn.RemoteAddr())
				return
			}
			logs.Error(i18n("log_err_reading_conn"), err) //读取连接时发生错误
			continue
		}
		if n == _ReedBufSize { //如果缓冲区满
			//if rbuf[n-1] == '\n' || rbuf[n-1] == '\r' { //有结束符
			//	logs.Debug(i18n("log_debug_terminator_received")) //接收到结束符
			//	isend = true                                      //设置结束标志
			//	rbuf = rbuf[:n-1]                                 //删除掉结束符
			//	n -= 1
			//} else { //数据区满，但没有接收到结束符
			isend = false
			go func(nc net.Conn, done chan bool) {
				n := 10
				cnt := make(chan int, n)
				for i := 0; i < n; i++ {
					cnt <- i
				}
				defer close(cnt)
				for {
					select {
					case c := <-cnt:
						time.Sleep(1 * time.Millisecond)
						if c == 9 {
							isend = true
							h.processData(buf)
							buf = buf[:0]
							return
						}
					case <-done:
						return
					}
				}
			}(h.Conn, read_start)
			//}
			buf = append(buf, rbuf[:n]...) //保存临时缓冲区数据到全局缓冲区
		} else { //缓冲区未满
			isend = true //
			//if rbuf[n-1] == '\n' || rbuf[n-1] == '\r' { //有结束符
			//	logs.Debug(i18n("log_debug_terminator_received")) //接收到结束符
			//	rbuf = rbuf[:n-1]                                 //删除掉结束符
			//	n -= 1
			//}
			buf = append(buf, rbuf[:n]...)
		}
		if isend { //接收数据结束
			h.processData(buf)
			buf = buf[:0]
		}
	}
}

//处理接收到的数据
func (h *ConnHandel) processData(buf []byte) error {
	var respData interface{}
	var resp RespMsg
	rn := len(buf)
	h.RxBytes += int64(rn) //接收字节数统计
	h.RxTimes += 1         //接收次数统计
	if rn < 4 {            //报文不能小于4字节
		resp.Ok = false
		resp.Msg = "comm_less_4byte"
		resp.Data = i18n(resp.Msg)
		respData = resp
	} else { //报文长度合格
		hex_data, ok := crc16.BytesCheckCRC(buf) //CRC16校验
		if !ok {                                 //校验不正确
			msg := i18n("log_err_check_code")
			logs.Error(msg)
			resp.Ok = false
			resp.Msg = "log_err_check_code"
			resp.Data = msg
			respData = resp
		} else { //校验通过
			fc, _, data, err := splitData(hex_data) //分割数据,提取功能码和数据
			if err != nil {
				logs.Error(i18n("log_err_splite_data"), err.Error())
			}
			//logs.Debug("功能码:%X,保留字节:%X,数据内容:%s", fc, hex_data[1], string(data))
			respData = h.functionExec(fc, data) //执行功能码,返回数据
		}
	}
	respbytes, err := json.Marshal(respData) //将结构数据转换为字节
	if err != nil {
		logs.Error(i18n("log_err_encode_json"), err.Error())
	}
	respbytes = encodeResponse(buf[0], buf[1], respbytes) //编码响应信息
	tn, err := h.Conn.Write(respbytes)                    //写响应
	if err != nil {
		logs.Error(i18n("log_err_socket_write"), err.Error())
		return err
	} else {
		h.TxBytes += int64(tn) //发送字节数统计
		h.TxTimes += 1         //发送次数统计
	}
	return nil
}

//执行功能码指定的功能
func (h *ConnHandel) functionExec(funcCode byte, data []byte) interface{} {
	var respData interface{}

	if funcCode == _FC_Ping { //连接测试
		return MakeRespMsg(true, fmt.Sprintf("MmKv:%s", VERSION), time.Now().Local().UnixMicro())
	}
	if !h.Logged && funcCode != _FC_Login { //尚未登录且非登录指令
		respData = MakeRespMsg(false, "user_unlogin", i18n("user_unlogin"))
	} else { //已经登录或者有登录指令
		switch funcCode {
		case _FC_ReadSingleKey: //读单个标签
			respData = h.ReadSingle(data)
		case _FC_ReadMultiKey: //读多个标签
			respData = h.ReadMulti(data)
		case _FC_WriteSingleKey, _FC_WriteMultiKey: //写单个或者写多个标签
			respData = h.Write(funcCode, data)
		case _FC_DeleteSingleKey: //删除单个标签
			respData = h.DeleteSingle(data)
		case _FC_DeleteMultiKey: //删除多个标签
			respData = h.DeleteMulti(data)
		case _FC_Login: //登录
			respData = h.Login(data)
		case _FC_SelfIncrease: //标签自增
			respData = h.SelfIncrease(data, 1)
		case _FC_SelfDecrease: //标签自减
			respData = h.SelfIncrease(data, -1)
		case _FC_PipePush: //压入管道
			respData = h.PipePush(data)
		case _FC_PipeFiFoPull, _FC_PipeFiLoPull: //从管道拉取数据
			respData = h.PipePull(funcCode, data)
		case _FC_PipeLenght: //获取管道长度
			respData = h.PipeLength(data)
		case _FC_GetKeys: //获取已经存在的所有键
			respData = h.GetKeys()
		case _FC_GetUsers: //获取已经存在的所有用户名
			respData = h.GetUsers()
		default: //未定义的功能码
			respData = MakeRespMsg(false, "fcode_undefined", fmt.Sprintf(i18n("fcode_undefined"), funcCode))
		}
	}
	if funcCode > 1 {
		fcmsg, ok := _FC_MAP[funcCode]
		if !ok {
			fcmsg = fmt.Sprintf("Undefined Function Code:%X", funcCode)
		}
		logs.Debug("%s -> %s : %s", h.Conn.RemoteAddr(), fcmsg, string(data))
	}
	return respData
}

//读取单个标签的值
func (h *ConnHandel) ReadSingle(data []byte) RespMsg {
	var resp RespMsg
	if key, err := decodeKey(data); err != nil {
		resp = MakeRespMsg(false, "ReadSingle.decodeKey", err.Error())
	} else {
		resp.Data, resp.Ok = Db.MmReadSingle(key)
		if !resp.Ok {
			resp.Data = fmt.Sprintf(i18n("undefined_key"), key)
		}
		resp.Msg = key
	}
	return resp
}

//读取多个标签的值
func (h *ConnHandel) ReadMulti(data []byte) RespMsg {
	if keys, err := decodeKeys(data); err != nil {
		return MakeRespMsg(false, "ReadMulti.decodeKeys", err.Error())
	} else {
		return MakeRespMsg(true, fmt.Sprintf("%d", len(keys)), Db.MmReadMulti(keys))
	}
}

//写标签的值
func (h *ConnHandel) Write(fc byte, data []byte) RespMsg {
	if kvs, err := decodeMap(data); err != nil {
		return MakeRespMsg(false, err.Error(), "")
	} else {
		data := Db.MmWriteMulti(kvs)
		if fc == _FC_WriteSingleKey { //写单个标签
			if len(data) > 0 { //有反馈数据
				return data[0]
			} else { //没有反馈数据
				return MakeRespMsg(false, "write_no_tag", i18n("write_no_tag"))
			}
		}
		if len(data) > 0 { //有反馈数据
			return MakeRespMsg(true, fmt.Sprintf("%d", len(data)), data)
		} else { //没有反馈数据
			return MakeRespMsg(false, "write_no_tag", i18n("write_no_tag"))
		}
	}
}

//删除单个标签的值
func (h *ConnHandel) DeleteSingle(data []byte) RespMsg {
	var resp RespMsg
	if key, err := decodeKey(data); err != nil {
		resp.Ok = false
		resp.Msg = "DeleteSingle.decodeKey"
		resp.Data = err.Error()
	} else {
		dels := Db.MmDeleteSingle(key)
		resp.Ok = true
		resp.Msg = "DeleteSingleKey"
		resp.Data = dels
	}
	return resp
}

//删除多个标签的值
func (h *ConnHandel) DeleteMulti(data []byte) RespMsg {
	var resp RespMsg
	if keys, err := decodeKeys(data); err != nil {
		resp.Ok = false
		resp.Msg = "DeleteMulti.decodeKeys"
		resp.Data = err.Error()
	} else {
		dels := Db.MmDeleteMulti(keys)
		resp.Ok = true
		resp.Msg = "DeleteMultiKey"
		resp.Data = dels
	}
	return resp
}

//删除多个标签的值
func (h *ConnHandel) Login(data []byte) RespMsg {
	var resp RespMsg
	if user, err := decodeUserMsg(data); err != nil {
		resp.Ok = false
		resp.Msg = "Login.decodeUserMsg"
		resp.Data = err.Error()
	} else {
		if ok, err := Db.MmCheckUser(user.Username, user.Password); ok {
			resp.Ok = true
			resp.Msg = VERSION
			resp.Data = user.Username
			h.Logged = true
			h.LogAt = time.Now()
			h.User = user.Username
			logs.Info(i18n("log_info_user_login"), user.Username, h.Conn.RemoteAddr())
		} else {
			resp.Ok = false
			resp.Msg = err.Error()
			resp.Data = i18n(resp.Msg)
		}
	}
	return resp
}

//标签自增
func (h *ConnHandel) SelfIncrease(data []byte, value int64) RespMsg {
	if key, err := decodeKey(data); err != nil {
		return MakeRespMsg(false, "SelfIncrease.decodeKey", err.Error())
	} else {
		newv, err := Db.MmSelfIncrease(key, value)
		if err != nil {
			return MakeRespMsg(false, "", err.Error())
		} else {
			return MakeRespMsg(true, key, newv)
		}
	}
}

//压入管道
func (h *ConnHandel) PipePush(data []byte) RespMsg {
	if kvs, err := decodeMap(data); err != nil {
		return MakeRespMsg(false, "PipePush.decodeMap", err.Error())
	} else {
		var n int = 0
		var err error
		var key string
		for k, v := range kvs {
			key = k
			n, err = Db.MmPipePush(k, v)
		}
		if err != nil { //有反馈数据
			return MakeRespMsg(false, "pipe push fail", err.Error())
		} else { //没有反馈数据
			return MakeRespMsg(true, key, n)
		}
	}
}

//从管道拉取数据
func (h *ConnHandel) PipePull(fc byte, data []byte) RespMsg {
	var resp RespMsg
	if key, err := decodeKey(data); err != nil {
		resp = MakeRespMsg(false, "PipePull.decodeKey", err.Error())
	} else {
		l, data, err := Db.MmPipePull(fc, key)
		if err != nil {
			resp = MakeRespMsg(false, fmt.Sprintf("%d", l), err.Error()) //无数据的时候长度为-1,其他错误时长度为0
		} else {
			resp = MakeRespMsg(true, fmt.Sprintf("%d", l), data)
		}
	}
	return resp
}

//获取管道长度
func (h *ConnHandel) PipeLength(data []byte) RespMsg {
	var resp RespMsg
	if key, err := decodeKey(data); err != nil {
		resp = MakeRespMsg(false, "PipeLength.decodeKey", err.Error())
	} else {
		l, err := Db.MmPipeLength(key)
		if err != nil {
			resp = MakeRespMsg(false, "get pipe length fail", err.Error()) //无数据的时候长度为-1,其他错误时长度为0
		} else {
			resp = MakeRespMsg(true, key, l)
		}
	}
	return resp
}

//获取现存的非系统keys
func (h *ConnHandel) GetKeys() RespMsg {
	var resp RespMsg
	var keys []string
	Db.Range(func(k, v interface{}) bool {
		key, ok := k.(string)
		if ok {
			keys = append(keys, key)
		}
		return true
	})
	resp.Data = keys
	resp.Ok = true
	return resp
}

//获取现存的系统users
func (h *ConnHandel) GetUsers() RespMsg {
	var resp RespMsg
	resp.Data = Db.MmGetUsersDict()
	resp.Ok = true
	return resp
}
