package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/bkzy-wangjp/mmkv/client"
)

var (
	MMKV_HOST = flag.String("h", "127.0.0.1", "mmkv内存数据库的主机地址")
	MMKV_PORT = flag.Int("p", 9646, "mmkv内存数据库的端口号")
	MMKV_USER = flag.String("u", "admin", "mmkv内存数据库的登录用户名")
	MMKV_PSWD = flag.String("psw", "admin123", "mmkv内存数据库的登录用户密码")
	MMKV_LANG = flag.String("lang", "zh-CN", "语言类型")
	MMKV_SIZE = flag.Int("size", 10, "mmkv数据库连接池大小")
)

func main() {
	flag.Parse()
	client.NewClient(map[string]interface{}{
		"hostname": *MMKV_HOST,
		"username": *MMKV_USER,
		"password": *MMKV_PSWD,
		"langtype": *MMKV_LANG,
		"port":     *MMKV_PORT,
		"size":     *MMKV_SIZE,
		"timeout":  3000,
		"max_sec":  60,
	})
	var cmd, key string
lable:
	for {
		var rst interface{}
		var sec float64
		var err error
		var msg string
		var pipelen int64

		fmt.Print("mmkv -> ")
		fmt.Scanf("%s %s\n", &cmd, &key)
		//fmt.Print("\nmmkv -> ")
		switch strings.ToLower(cmd) {
		case "p", "ping":
			rst, sec, err = client.ServerTime()
		case "si", "incr":
			rst, sec, err = client.SelfIncrease(key)
		case "sd", "decr":
			rst, sec, err = client.SelfDecrease(key)
		case "w", "write":
			kvs := strings.Split(key, "=")
			if len(kvs) == 2 {
				rst, sec, err = client.Write(kvs[0], kvs[1])
			} else {
				err = fmt.Errorf("键值对格式错误,应为:key=value")
			}
		case "writes":
			kvs := make(map[string]interface{})
			er := json.Unmarshal([]byte(key), &kvs)
			if err == nil {
				fmt.Println(kvs)
				rst, sec, err = client.Writes(kvs)
			} else {
				err = fmt.Errorf("数据解析失败:%s", er.Error())
			}
		case "push":
			kvs := strings.Split(key, "=")
			if len(kvs) == 2 {
				rst, sec, err = client.PipePush(kvs[0], kvs[1])
			} else {
				err = fmt.Errorf("键值对格式错误,应为:key=value")
			}
		case "fifo", "lilo", "filo", "lifo": //拉取
			pipelen, rst, sec, err = client.PipePull(key, cmd)
			if err == nil {
				msg = fmt.Sprintf("当前管道长度:%d", pipelen)
			}
		case "pa", "pipeall":
			rst, sec, err = client.PipeAll(key)
		case "pl", "pipelen":
			rst, sec, err = client.PipeLength(key)
		case "r", "read":
			rst, sec, err = client.Read(key)
		case "keys", "getkeys", "tags":
			rst, sec, err = client.GetKeys()
		case "users", "getusers":
			rst, sec, err = client.GetUsers()
		case "delete":
			ok := false
			ok, sec, err = client.Delete(key)
			if ok {
				rst = fmt.Sprintf("成功删除 %s", key)
			} else {
				rst = fmt.Sprintf("删除 %s 失败", key)
			}
		case "h", "help":
			hmsg := "    h或者help:输出帮助信息\n"
			hmsg += "    p或者ping:测试与数据库的连接，输出数据库的当前时间\n"
			hmsg += "    keys或者tags:输出数据库中所有的键值对的键\n"
			hmsg += "    w或者write key=value:往数据库中写入一个键值对，'key'为键,'value'为值\n"
			hmsg += "    writes {\"key1\":\"value1\",\"key2\":\"value2\"}:往数据库中写入多个键值对，Json格式\n"
			hmsg += "    r或者read key:读取数据库中的键值对，'key'为键\n"
			hmsg += "    d或者delete key:删除数据库中的键值对，'key'为键\n"
			hmsg += "    si或者incr key:让键的值增加1(值必须为数字类型)，'key'为键\n"
			hmsg += "    di或者decr key:让键的值减小1(值必须为数字类型)，'key'为键\n"
			hmsg += "    si或者incr key:让键的值增加1(值必须为数字类型)，'key'为键\n"
			hmsg += "    push key=value:往数据库管道型键值对中追加一个值，'key'为键,'value'为值\n"
			hmsg += "    fifo key:以FIFO(先进先出)的方式从管道型键值对中拉取一个值，'key'为键\n"
			hmsg += "    filo key:以FILO(先进后出)的方式从管道型键值对中拉取一个值，'key'为键\n"
			hmsg += "    lilo key:以LILO(后进后出)的方式从管道型键值对中拉取一个值(效果等同于FIFO)，'key'为键\n"
			hmsg += "    lifo key:以LIFO(后进先出)的方式从管道型键值对中拉取一个值(效果等同于FILO)，'key'为键\n"
			hmsg += "    pa或者pipeall key:获取管道中的所有数据并清空管道，'key'为键\n"
			hmsg += "    pipelen key:检查管道的长度，'key'为键\n"
			hmsg += "    users:列出当前数据库中的账户名称\n"
			hmsg += "    q或者quit或者exit:退出数据库客户端\n"
			rst = hmsg
		case "q", "quit", "exit":
			break lable
		default:
			err = fmt.Errorf("无效指令")
		}

		if err != nil {
			fmt.Println(err.Error())
		} else {
			if strings.ToLower(cmd) == "h" || strings.ToLower(cmd) == "help" {
				fmt.Print(rst)
			} else {
				if msg == "" {
					fmt.Printf("执行耗时:%f秒\n", sec)
				} else {
					fmt.Printf("执行耗时:%f秒,%s\n", sec, msg)
				}
				StructFormatPrint(rst, "执行结果")
			}
		}
		fmt.Println()
	}
}

//Json格式打印输出结构体
func StructFormatPrint(msg interface{}, names ...string) {
	bs, _ := json.Marshal(msg)
	var out bytes.Buffer
	json.Indent(&out, bs, "", "\t")

	name := ""
	if len(names) > 0 {
		name = names[0]
	}
	if len(name) > 0 {
		name += ":"
	}
	fmt.Printf("%s%v\n", name, out.String())
}
