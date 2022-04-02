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
