package mmdb

import (
	"log"
	"net"
	"sync"
)

const (
	VERSION      = "v0.0.1"    //版本号
	_ReedBufSize = 1024 * 1024 //读信息通道缓存字节数
)

//运行数据库结构
type MemoryDb struct {
	sync.Map
}

//var db MemoryDb

func Run(address string) error {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("accept error: ", err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(c net.Conn) {
	defer c.Close()
	for {
		//read from the connection
		var buf = make([]byte, _ReedBufSize)
		_, err := c.Read(buf)
		if err != nil {
			log.Println("conn read error:", err)
			return
		}
	}
}
