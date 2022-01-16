package client

import (
	"testing"
)

func TestRun(t *testing.T) {
	params := map[string]interface{}{"port": 8888, "size": 10}
	pool, err := NewConnPool(params)
	if err != nil {
		t.Error(err)
	} else {
		for i, h := range pool.Conns {
			t.Logf("No. %d 连接:%v", i, h)
		}
	}
}

func TestPing(t *testing.T) {
	username := "admin"
	password := "admin123"
	address := "127.0.0.1:8888"
	h, err := newConneHandle(username, password, address)
	if err != nil {
		t.Errorf("建立连接时错误:%s", err.Error())
	} else {
		sec, err := h.ping()
		if err != nil {
			t.Errorf("PING时错误:%s", err.Error())
		} else {
			t.Logf("PING接口耗时间 %f 秒", sec)
		}
	}
}

func TestWrite(t *testing.T) {
	data := map[string]interface{}{
		"foo":   "bar",
		"张三":    93.5,
		"China": "五星红旗",
		"美国":    "星条旗",
	}
	params := map[string]interface{}{"port": 8888, "size": 50}
	pool, err := NewConnPool(params)
	if err != nil {
		t.Error(err)
	} else {
		for i := 0; i < 10; i++ {
			_, _, err := pool.Write(data)
			if err != nil {
				t.Error(err)
			}
		}
	}
}

func TestRead(t *testing.T) {
	data := []string{
		"foo",
		"张三",
		"China",
		"美国",
	}
	params := map[string]interface{}{"port": 8888, "size": 50}
	pool, err := NewConnPool(params)
	if err != nil {
		t.Error(err)
	} else {
		for i := 0; i < 2; i++ {
			msg, sec, err := pool.Read(data...)
			if err != nil {
				t.Error(err)
			} else {
				t.Log(msg, sec)
			}
		}
	}
}

func TestDelete(t *testing.T) {
	data := []string{
		"foo",
		"张三",
		"China",
		"美国",
	}
	params := map[string]interface{}{"port": 8888, "size": 50}
	pool, err := NewConnPool(params)
	if err != nil {
		t.Error(err)
	} else {
		for i := 0; i < 10; i++ {
			msg, sec, err := pool.Delete(data...)
			if err != nil {
				t.Error(err)
			} else {
				t.Log(msg, sec)
			}
		}
	}
}
