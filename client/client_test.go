package client

import (
	"testing"
)

func TestRun(t *testing.T) {
	params := map[string]interface{}{"port": 9646, "size": 10}
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
	address := "127.0.0.1:9646"
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
	params := map[string]interface{}{"port": 9646, "size": 50}
	pool, err := NewConnPool(params)
	if err != nil {
		t.Error(err)
	} else {
		for i := 0; i < 10; i++ {
			msg, sec, err := pool.Writes(data)
			if err != nil {
				t.Error(err)
			} else {
				t.Log(msg, sec)
			}
		}
	}
}

func TestReadMulti(t *testing.T) {
	data := []string{
		"foo",
		"张三",
		"China",
		"美国",
	}
	params := map[string]interface{}{"port": 9646, "size": 50}
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

func TestReadSingle(t *testing.T) {
	data := []string{
		"foo",
		"张三",
		"China",
		"美国",
	}
	params := map[string]interface{}{"port": 9646, "size": 50}
	pool, err := NewConnPool(params)
	if err != nil {
		t.Error(err)
	} else {
		for _, key := range data {
			msg, sec, err := pool.Read(key)
			if err != nil {
				t.Error(err)
			} else {
				t.Log(key, msg, sec)
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
	params := map[string]interface{}{"port": 9646, "size": 50}
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

func TestSelfAdd(t *testing.T) {
	Keys := []string{
		"foo",
		"张三",
		"China",
		"美国",
	}
	params := map[string]interface{}{"port": 9646, "size": 50}
	pool, err := NewConnPool(params)
	if err != nil {
		t.Error(err)
	} else {
		for i := 0; i < 10; i++ {
			for _, k := range Keys {
				msg, sec, err := pool.SelfIncrease(k)
				if err != nil {
					t.Error(err)
				} else {
					t.Log(msg, sec)
				}
			}
		}
		for i := 0; i < 10; i++ {
			for _, k := range Keys {
				msg, sec, err := pool.SelfDecrease(k)
				if err != nil {
					t.Error(err)
				} else {
					t.Log(msg, sec)
				}
			}
		}
	}
}

func TestPipe(t *testing.T) {
	Keys := []string{
		"foo",
		"张三",
		"China",
		"美国",
	}
	params := map[string]interface{}{"port": 9646, "size": 50}
	pool, err := NewConnPool(params)
	if err != nil {
		t.Error(err)
	} else {
		t.Log("--------检查变量------------")
		for _, k := range Keys {
			msg, sec, err := pool.Read(k)
			if err != nil {
				t.Error(k, err)
			} else {
				t.Log(k, msg, sec)
			}
		}
		//压入管道
		t.Log("--------压入数据------------")
		for i := 0; i < 10; i++ {
			for j, k := range Keys {
				msg, sec, err := pool.PipePush(k, i+j+1)
				if err != nil {
					t.Error(k, err)
				} else {
					t.Log(k, msg, sec)
				}
			}
		}
		//观察
		t.Log("--------读取检查------------")
		for _, k := range Keys {
			msg, sec, err := pool.Read(k)
			if err != nil {
				t.Error(k, err)
			} else {
				t.Log(k, msg, sec)
			}
		}
		//fifo拉取
		t.Log("--------FIFO拉取------------")
		for _, k := range Keys {
			length, msg, sec, err := pool.PipePull(k, "fifo")
			if err != nil {
				t.Error(k, err)
			} else {
				t.Log(k, length, msg, sec)
			}
		}
		t.Log("--------再次检查变量------------")
		for _, k := range Keys {
			msg, sec, err := pool.Read(k)
			if err != nil {
				t.Error(k, err)
			} else {
				t.Log(k, msg, sec)
			}
		}
		//filo拉取
		t.Log("--------FILO拉取------------")
		for _, k := range Keys {
			length, msg, sec, err := pool.PipePull(k, "filo")
			if err != nil {
				t.Error(k, err)
			} else {
				t.Log(k, length, msg, sec)
			}
		}
		t.Log("--------再次检查变量------------")
		for _, k := range Keys {
			msg, sec, err := pool.Read(k)
			if err != nil {
				t.Error(k, err)
			} else {
				t.Log(k, msg, sec)
			}
		}
		t.Log("--------检查剩余长度------------")
		for _, k := range Keys {
			msg, sec, err := pool.PipeLength(k)
			if err != nil {
				t.Error(k, err)
			} else {
				t.Log(k, msg, sec)
			}
		}
	}
}

func TestGetKeys(t *testing.T) {
	params := map[string]interface{}{"port": 9646, "size": 5}
	pool, err := NewConnPool(params)
	if err != nil {
		t.Error(err)
	} else {
		keys, sec, err := pool.GetKeys()
		if err != nil {
			t.Error(err)
		} else {
			t.Log(sec)
			for k, v := range keys {
				t.Log(k, v)
			}
		}
	}
}

func TestGetUsers(t *testing.T) {
	params := map[string]interface{}{"port": 9646, "size": 5}
	pool, err := NewConnPool(params)
	if err != nil {
		t.Error(err)
	} else {
		users, sec, err := pool.GetUsers()
		if err != nil {
			t.Error(err)
		} else {
			t.Log(sec)
			for k, v := range users {
				t.Log(k, v)
			}
		}
	}
}
