package client

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	crc16 "github.com/bkzy-wangjp/CRC16"
)

//一个通讯连接
type ConnHandel struct {
	Id           int64     //序号
	Conn         net.Conn  //通讯连接
	timeout      int64     //超时时间,ms,0无超时限制
	ConnServer   bool      //与服务器已建立连接
	Logged       bool      //是否已登录标志
	Working      bool      //工作中
	retry        bool      //是否已经重试发送
	TxBytes      int64     //发送字节数
	RxBytes      int64     //接收字节数
	TxTimes      int64     //发送次数
	RxTimes      int64     //接收次数
	lastFc       byte      //最后一次写数据的功能码
	lastRev      byte      //最后一次写数据的识别码
	WorkeAt      time.Time //本次开始工作的时间
	LoginAt      time.Time //登录时间
	CreatedAt    time.Time //创建时间
	LoseServerAt time.Time //最近一次连接不上服务器的时间
}

//新建通道连接句柄
func newConneHandle(username, password, address string) (*ConnHandel, error) {
	handle := new(ConnHandel)
	handle.CreatedAt = time.Now()
	err := handle.dial(address) //建立连接
	if err != nil {
		return nil, err
	}
	err = handle.login(username, password) //登录
	if err != nil {
		return nil, err
	}
	return handle, nil
}

//建立与服务器的连接
func (h *ConnHandel) dial(addr string) error {
	var err error
	if h.timeout == 0 {
		h.Conn, err = net.Dial("tcp", addr)
	} else {
		h.Conn, err = net.DialTimeout("tcp", addr, time.Duration(h.timeout)*time.Microsecond)
	}
	if err != nil {
		h.ConnServer = false
	} else {
		h.ConnServer = true
	}
	return err
}

//建立与服务器的连接
func (h *ConnHandel) ping() (float64, error) {
	var data []byte
	st := time.Now()
	msg, err := h.write(_FC_Ping, data)
	if err != nil {
		return 0, err
	}
	if !msg.Ok {
		return 0, fmt.Errorf("%s", msg.Data)
	}
	return time.Since(st).Seconds(), nil
}

//登录
//登录不成功返回错误信息，登录成功无错误
func (h *ConnHandel) login(username, password string) error {
	user := new(UserMsg)
	user.Username = username
	user.Password = password
	msgdata, err := json.Marshal(user) //将结构数据转换为字节
	if err != nil {
		return fmt.Errorf(i18n("log_err_user_marshal"), err.Error())
	}
	msg, err := h.write(_FC_Login, msgdata)
	if err != nil {
		return err
	}
	if !msg.Ok {
		return fmt.Errorf("%s", msg.Data)
	}
	h.Logged = true
	h.LoginAt = time.Now()
	return nil
}

//写请求数据并返回响应信息
func (h *ConnHandel) write(fc byte, data []byte) (*RespMsg, error) {
	h.lastFc = fc
	h.lastRev += 1
	reqmsg := encodeRequest(h.lastFc, h.lastRev, data)
	n, err := h.Conn.Write(reqmsg)
	if err != nil {
		return nil, fmt.Errorf(i18n("log_err_write_faile"), err.Error())
	}
	if n != len(reqmsg) {
		return nil, fmt.Errorf(i18n("log_err_write_len_no_mantch"))
	}
	h.TxBytes += int64(n) //发送字节数统计
	h.TxTimes += 1        //发送次数统计

	respmsg, err := h.read()
	if err != nil && !h.retry { //检查到错误有一次重发的机会
		h.retry = true
		return h.write(fc, data)
	}
	if h.retry {
		h.retry = false
	}
	return respmsg, err
}

//读取数据,返回的接口是反馈信息的data部分或者ok部分
//返回接收到的 RespMsg 和错误信息
func (h *ConnHandel) read() (*RespMsg, error) {
	var buf []byte
	var crccheckok bool
	isend := false //是否读取结束
	for {
		//read from the connection
		var rbuf = make([]byte, _ReedBufSize)
		n, er := h.Conn.Read(rbuf) //读取缓冲通道，如果无数据则挂起
		if er != nil {
			h.ConnServer = false
			err := fmt.Errorf(i18n("log_err_reading_conn"), er) //读取连接时发生错误
			return nil, err
		}
		if n == _ReedBufSize { //如果缓冲区满
			buf = append(buf, rbuf[:n]...)           //保存临时缓冲区数据到全局缓冲区
			_, crccheckok = crc16.BytesCheckCRC(buf) //CRC16校验
			if crccheckok {
				isend = true
			} else {
				isend = false
			}
		} else { //缓冲区未满
			isend = true //
			buf = append(buf, rbuf[:n]...)
		}
		if isend { //接收数据结束
			if len(buf) < 4 { //数据长度不够
				buf = buf[:0]
				isend = false
			} else {
				break
			}
		}
	}
	return h.processData(buf, crccheckok)
}

//处理接收到的数据
//返回接收到的 RespMsg 和错误信息
func (h *ConnHandel) processData(buf []byte, crcchecked bool) (*RespMsg, error) {
	rn := len(buf)
	h.RxBytes += int64(rn) //接收字节数统计
	h.RxTimes += 1         //接收次数统计
	hex_data := buf[:len(buf)-2]
	if !crcchecked { //未执行过CRC校验
		data, ok := crc16.BytesCheckCRC(buf) //CRC16校验
		if !ok {                             //校验不正确
			err := fmt.Errorf(i18n("log_err_check_code"))
			return nil, err
		} else {
			hex_data = data
		}
	}
	fc, rev, data, err := splitData(hex_data) //分割数据,提取功能码和数据
	if err != nil {
		err := fmt.Errorf(i18n("log_err_splite_data"), err.Error())
		return nil, err
	}
	if fc != h.lastFc || rev != h.lastRev { //校验与发送记录是否匹配
		err = fmt.Errorf(i18n("log_err_read_write_no_match"))
		return nil, err
	}
	//logs.Debug("功能码:%X,保留字节:%X,数据内容:%s", fc, rev, string(data))
	return decodeResponse(data) //执行功能码,返回数据
}

//解析接收到的数据
/*
func (h *ConnHandel) decodeData(data []byte) (str_msg string, ifc_data interface{}, err error) {
	funcCode := h.lastFc
	respmsg, er := decodeResponse(data) //解析反馈数据
	if er != nil {
		err = er
		return
	}
	var msgs []RespMsg
	msg, ok := respmsg.(RespMsg) //按照单个信息方式解析
	if !ok {
		msgs, _ = respmsg.([]RespMsg)
	} else {
		if !msg.Ok {
			err = fmt.Errorf("%s", msg.Data) //按照多个信息方式解析
			return
		}
	}

	switch funcCode {
	case _FC_ReadSingleKey:
		ifc_data = msg.Data
	case _FC_ReadMultiKey:
		ifc_data = msgs
	case _FC_WriteSingleKey, _FC_WriteMultiKey:
		ifc_data = true
	case _FC_DeleteSingleKey:
		ifc_data = true
	case _FC_DeleteMultiKey:
		ifc_data = true
	case _FC_Login:
		ifc_data = true
		h.Logged = true
		h.LoginAt = time.Now()
	case _FC_Ping:
		ifc_data = msg.Data
	default:
		ifc_data = msg.Data
	}
	return
}
*/
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

//编码请求报文
func encodeRequest(funcCode, reverse byte, data []byte) []byte {
	var resp []byte
	resp = append(resp, funcCode, reverse)
	resp = append(resp, data...)
	return crc16.BytesAndCrcSum(resp, true)
}

//解码信息反馈报文
func decodeResponse(hex_data []byte) (*RespMsg, error) {
	msg := new(RespMsg)
	err := json.Unmarshal(hex_data, msg)
	/*
		if err != nil {
			var msgs []RespMsg
			err = json.Unmarshal(hex_data, &msgs)
			if err == nil {
				msgdata = msgs
			}
		} else {
			msgdata = msg
		}
	*/
	return msg, err
}

//自由请求接口
//请求参数:
// funcCode byte : 功能码
// msg interface{} : 符合json结构和功能码要求的信息
//反回参数：
// *RespMsg : 服务器返回的数据
// float64 : 执行耗时,单位秒
// error : 错误信息(已经判断是否ok)
func (h *ConnHandel) request(funcCode byte, msg interface{}) (*RespMsg, float64, error) {
	st := time.Now()
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, 0, err
	}
	resp, err := h.write(funcCode, data)
	if err != nil {
		return nil, time.Since(st).Seconds(), err
	}
	if !resp.Ok {
		return nil, time.Since(st).Seconds(), fmt.Errorf("%s", resp.Data)
	}
	return resp, time.Since(st).Seconds(), nil
}
