# mmkv

一种基于内存的键值对数据库，类似于Redis。

a Memory Key-Value database, like Redis.

## 服务器端

```go
//启动内存KV值数据库
package main
import "micollect/models/mmkv"
func main() {
    go mmkv.Run(map[string]string{user_name: pswd},
     map[string]interface{}{"port": port, "ip": ip, "languige": LANG_TYPE})
}

```

## 请求报文格式

|功能码|预留字节|数据载荷|CRC16_H|CRC16_L|功能说明|
|------|-------|--------|----|----|----|
|0x02 |0x00|`{"username":"admin","password":"admin123"}`|0x22|0x7D|登录示例|

>上表中数据最终的HEX报文为：`02 00 7B 22 75 73 65 72 6E 61 6D 65 22 3A 22 61 64 6D 69 6E 22 2C 22 70 61 73 73 77 6F 72 64 22 3A 22 61 64 6D 69 6E 31 32 33 22 7D`

- 报文以HEX(16进制)编码发送

- 报文的前两个字节是报文头(header),第一字节为功能码，第二字节为系统保留码

- 报文的最后两个字节为CRC16校验码，高字节在前，低字节在后，根据报文头和报文载荷自动计算

- 报文头和校验码之间的为报文载荷(payload), 报文载荷为UTF8格式的JSON数据转换而来

## 返回报文格式

|功能码|预留字节|数据载荷|CRC16_H|CRC16_L|功能说明|
|------|-------|--------|----|----|----|
|0x01 |0x00|`{"ok":true,"msg":"MmKv:v0.0.1","data":1646100208058699}`|0xB4|0xC2|连接测试示例|

>上表中数据最终的HEX报文为：`01 00 7B 22 6F 6B 22 3A 74 72 75 65 2C 22 6D 73 67 22 3A 22 4D 6D 4B 76 3A 76 30 2E 30 2E 31 22 2C 22 64 61 74 61 22 3A 31 36 34 36 31 30 30 32 30 38 30 35 38 36 39 39 7D B4 C2`

- 报文以HEX(16进制)编码发送

- 报文的前两个字节是报文头(header),第一字节为功能码，第二字节为系统保留码。返回报文的报文头与请求报文的报文头要完全一致

- 报文的最后两个字节为CRC16校验码，高字节在前，低字节在后，根据报文头和报文载荷自动计算

- 报文头和校验码之间的为报文载荷(payload), 报文载荷为UTF8格式的JSON数据转换而来

### 返回报文载荷的说明

返回报文的数据载荷的JSON格式是固定的，如下表所示：

|键(Key)|数据格式|说明|
|-------|--------|----|
|`ok`   |bool    |返回数据的状态， 如果正确， 其值为`true`， 如果不正确， 其值为`false`|
|`msg`  |string  |当`ok==true`时，此处为返回数据的附加信息; 当`ok==false`时，该值为错误信息|
|`data` |struct  |不定结构，根据请求内容不同返回不同的内容; 当`ok==false`时，该值为错误的详细描述|

## 功能码说明及示例

|功能码|功能说明|请求载荷JSON|返回载荷JSON|
|------|-------|--------|------|
|0x01    |连接测试|无|`{"ok":true, "msg":"MmKv:v0.0.1", "data":1646099841652811}`|
|0x02    |用户登录|`{"username":"admin", "password":"admin123"}`| `{"ok":true, "msg":"v0.0.1", "data":"admin"}`|
|0x03    |写单个标签|`{"tag_a":12345}`|`{"ok":true, "msg":"tag_a", "data":12345}`|
|0x04    |写多个标签|`{"tag_a":12345, "tag_b":"abcdefg"}`|`{"ok":true,"msg":"", "data":[{"ok":true, "msg":"tag_a", "data":12345}, {"ok":true, "msg":"tag_b", "data":"abcdefg"}]}`|
|0x05    |读取单个标签|`"tag_a"`|`{"ok":true, "msg":"tag_a", "data":12345}`|
|0x06    |读取多个标签|`["tag_a","tag_b"]`|`{"ok":true,"msg":"", "data":[{"ok":true, "msg":"tag_a", "data":12345}, {"ok":true, "msg":"tag_b", "data":"abcdefg"}]}`|
|0x07    |删除单个标签|`"tag_a"`|`{"ok":true, "msg":"DeleteSingleKey", "data":0}`|
|0x08    |删除多个标签|`["tag_a","tag_b"]`|`{"ok":true,"msg":"DeleteMultiKey","data":0}`|
|0x09    |数据自增|`"tag_a"`|`{"ok":true,"msg":"tag_a","data":2}`|
|0x0A    |数据自减|`"tag_a"`|`{"ok":true,"msg":"tag_a","data":2}`|
|0x0B    |压入管道|`{"tag_pipe":123456.789}`|`{"ok":true,"msg":"tag_pipe","data":4}`|
|0x0C    |从管道拉取(FIFO)|`"tag_pipe"`|`{"ok":true,"msg":"3","data":123456.789}`|
|0x0D    |从管道拉取(FILO)|`"tag_pipe"`|`{"ok":true,"msg":"2","data":123456.789}`|
|0x0E    |获取管道当前长度 |`"tag_pipe"`|`{"ok":true,"msg":"tag_pipe","data":2}`|
