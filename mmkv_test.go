package mmkv

import (
	"encoding/hex"
	"reflect"
	"testing"
)

func TestDecodeKey(t *testing.T) {
	tests := []struct {
		hex_data string
		key      string
	}{
		{"226b65793122", "key1"},
		{"22e4bda0e5a5bde4b896e7958c22", "你好世界"},
		{"0525225379732e46657272794d616e2e53746174757322be3c0126fa81", ""},
	}
	for _, tt := range tests {
		bt, err := hex.DecodeString(tt.hex_data)
		if err != nil {
			t.Errorf("hex字符串转换为字节出错:%s", err.Error())
		} else {
			key, err := decodeKey(bt)
			if err != nil {
				t.Errorf("decode错误:%s", err.Error())
			} else {
				if key != tt.key {
					t.Errorf("错误,期望值是:%s,得到的值是:%s", tt.key, key)
				}
			}
		}
	}
}

func TestDecodeKeys(t *testing.T) {
	tests := []struct {
		hex_datas string
		keys      []string
	}{
		{"5b226b657931222c22e4bda0e5a5bde4b896e7958c225d", //["key1","你好世界"]
			[]string{"key1", "你好世界"}},
		{"5b226b657931225d", //["key1"]
			[]string{"key1"}},
	}
	for _, tt := range tests {
		bt, err := hex.DecodeString(tt.hex_datas)
		if err != nil {
			t.Errorf("hex字符串转换为字节出错:%s", err.Error())
		} else {
			keys, err := decodeKeys(bt)
			if err != nil {
				t.Errorf("decode错误:%s", err.Error())
			} else {
				for i, k := range keys {
					if k != tt.keys[i] {
						t.Errorf("错误,期望值是:%s,得到的值是:%s", tt.keys[i], k)
					}
				}
			}
		}
	}
}

func TestDecodeMap(t *testing.T) {
	tests := []struct {
		hex_datas string
		kvs       map[string]interface{}
	}{
		{"7b226b657931223a31323334357d", //{"key1":12345}
			map[string]interface{}{"key1": 12345, "key2": "string", "key3": []byte{123, 25, 32, 34, 45}},
		},
		{"7b226b657932223a22737472696e67227d", //{"key2":"string"}
			map[string]interface{}{"key2": "string"},
		},
		{"7b226b657933223a5b3132332c32352c33322c33342c34355d7d", //{"key3":[123,25,32,34,45]}
			map[string]interface{}{"key3": []byte{123, 25, 32, 34, 45}},
		},
		{"7b226b657931223a31323334352c226b657932223a22737472696e67222c226b657933223a5b3132332c32352c33322c33342c34355d7d", //{"key1":12345,"key2":"string","key3":[123,25,32,34,45]}
			map[string]interface{}{"key1": 12345, "key2": "string", "key3": []byte{123, 25, 32, 34, 45}},
		},
	}
	for _, tt := range tests {
		bt, err := hex.DecodeString(tt.hex_datas)
		if err != nil {
			t.Errorf("hex字符串转换为字节出错:%s", err.Error())
		} else {
			kvs, err := decodeMap(bt)
			if err != nil {
				t.Errorf("decode错误:%s", err.Error())
			} else {
				for k, v := range kvs {
					if value, ok := tt.kvs[k]; !ok {
						if value != v {
							t.Errorf("错误,%s的期望值是:%s,得到的值是:%s", k, tt.kvs[k], v)
						}
					}
				}
			}
		}
	}
}

func TestReflect(t *testing.T) {
	tests := []interface{}{
		123,
		123.456,
		true,
		[]int{123, 25, 32, 34, 45},
		"abcd",
	}
	for _, v := range tests {
		t.Log(reflect.TypeOf(v))
		val, ok := v.(float64)
		t.Log(val, ok)
	}
}

func TestSelfAdd(t *testing.T) {
	tests := []interface{}{
		123,
		123.456,
		true,
		[]int{123, 25, 32, 34, 45},
		"abcd",
	}
	for _, v := range tests {
		rel, err := selfAdd(v, 1)
		if err != nil {
			t.Error(err.Error())
		} else {
			t.Log(rel)
		}
	}
}

func TestCrcCheck(t *testing.T) {
	tests := []struct {
		hex_data string
	}{
		{"0525225379732e46657272794d616e2e53746174757322be3c0126fa81"},
		{"05a2225379732e46657272794d616e2e53746174757322308201a35940"},
		{"02577b22757365726e616d65223a2261646d696e222c2270617373776f7264223a223031393230323361376262643733323530353136663036396466313862353030227ddf5e03587b225461672e4d542e404d616368696e65436f6465223a7b2254696d65223a22323032332d30342d32335432313a30393a33302e383438373135352b30383a3030222c2256616c7565223a302c2256616c7565537472223a226e4a4b524d5656746c68706d6b7455546a6e3644475a324f5a7179636a6b6b755450535568376e6151676d796b616b44472b51464b34614962486271792f6a4153343654744f6a704b6f646f4a624b6c396e4d7a44413d3d222c22506572696f64223a302c2244617461537461747573223a302c2244796e616d6963537461747573223a302c224572724d7367223a22227d7d54be"},
		{"22e4bda0e5a5bde4b896e7958c22"},
	}
	for _, tt := range tests {
		bt, err := hex.DecodeString(tt.hex_data)
		if err != nil {
			t.Errorf("hex字符串转换为字节出错:%s", err.Error())
		} else {
			t.Log(bt)
			key, err := decodeKey(bt)
			if err != nil {
				t.Errorf("decode错误:%s", err.Error())
			} else {
				t.Log(key)
			}
		}
	}
}
