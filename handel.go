package mmdb

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
			if err == io.EOF || strings.Contains(err.Error(), "close") { //客户端关闭
				logs.Info(i18n("log_info_client_shutdown"), h.Conn.RemoteAddr())
				return
			}
			logs.Error(i18n("log_err_reading_conn"), err) //读取连接时发生错误
			continue
		}
		if n == _ReedBufSize { //如果缓冲区满
			if rbuf[n-1] == '\n' || rbuf[n-1] == '\r' { //有结束符
				logs.Debug(i18n("log_debug_terminator_received")) //接收到结束符
				isend = true                                      //设置结束标志
				rbuf = rbuf[:n-1]                                 //删除掉结束符
				n -= 1
			} else { //数据区满，但没有接收到结束符
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
			}
			buf = append(buf, rbuf[:n]...) //保存临时缓冲区数据到全局缓冲区
		} else { //缓冲区未满
			isend = true                                //
			if rbuf[n-1] == '\n' || rbuf[n-1] == '\r' { //有结束符
				logs.Debug(i18n("log_debug_terminator_received")) //接收到结束符
				rbuf = rbuf[:n-1]                                 //删除掉结束符
				n -= 1
			}
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
	//hex_string := hex.EncodeToString(buf)
	//byte_data, er := hex.DecodeString(string(hex_string))
	//logs.Debug("hex_string:%s,byte_data:%s,错误信息:%v", hex_string, string(byte_data), er)
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
			logs.Debug("功能码:%X,保留字节:%X,数据内容:%s", fc, hex_data[1], string(data))
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
	if funcCode >= 48 { //ASCII码模式
		funcCode -= 48
	}
	if !h.Logged && funcCode != _FC_Login { //尚未登录且非登录指令
		var resp RespMsg
		resp.Ok = false
		resp.Msg = "user_unlogin"
		resp.Data = i18n(resp.Msg)
		respData = resp
	} else { //已经登录或者有登录指令
		switch funcCode {
		case _FC_ReadSingleKey:
			var resp RespMsg
			if key, err := decodeKey(data); err != nil {
				resp.Ok = false
				resp.Msg = err.Error()
				resp.Data = ""
			} else {
				resp.Data, resp.Ok = MmDb.MmReadSingle(key)
				if !resp.Ok {
					resp.Data = ""
				}
				resp.Msg = key
			}
			respData = resp
		case _FC_ReadMultiKey:
			var resp RespMsg
			if keys, err := decodeKeys(data); err != nil {
				resp.Ok = false
				resp.Msg = err.Error()
				resp.Data = ""
				respData = resp
			} else {
				respData = MmDb.MmReadMulti(keys)
			}
		case _FC_WriteSingleKey, _FC_WriteMultiKey:
			var resp RespMsg
			if kvs, err := decodeMap(data); err != nil {
				resp.Ok = false
				resp.Msg = err.Error()
				resp.Data = ""
			} else {
				MmDb.MmWriteMulti(kvs)
				resp.Ok = true
				if funcCode == _FC_WriteSingleKey {
					resp.Msg = "WriteSingleKey"
				} else {
					resp.Msg = "WriteMultiKey"
				}
				resp.Data = len(kvs)
			}
			respData = resp
		case _FC_DeleteSingleKey:
			var resp RespMsg
			if key, err := decodeKey(data); err != nil {
				resp.Ok = false
				resp.Msg = err.Error()
			} else {
				MmDb.MmDeleteSingle(key)
				resp.Ok = true
				resp.Msg = "DeleteSingleKey"
				resp.Data = 1
			}
			respData = resp
		case _FC_DeleteMultiKey:
			var resp RespMsg
			if keys, err := decodeKeys(data); err != nil {
				resp.Ok = false
				resp.Msg = err.Error()
			} else {
				MmDb.MmDeleteMulti(keys)
				resp.Ok = true
				resp.Msg = "DeleteMultiKey"
				resp.Data = len(keys)
			}
			respData = resp
		case _FC_Login:
			var resp RespMsg
			if user, err := decodeUserMsg(data); err != nil {
				resp.Ok = false
				resp.Msg = err.Error()
				resp.Data = ""
			} else {
				if ok, err := MmDb.MmCheckUser(user.Username, user.Password); ok {
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
			respData = resp
		case _FC_Ping:
			var resp RespMsg
			resp.Ok = true
			resp.Msg = "Ping connection"
			resp.Data = fmt.Sprint(time.Now().Local().UnixMicro())
			respData = resp
		default:
			var resp RespMsg
			resp.Ok = false
			resp.Msg = "fcode_undefined"
			resp.Data = fmt.Sprintf(i18n(resp.Msg), funcCode)
			respData = resp
		}
	}
	return respData
}
