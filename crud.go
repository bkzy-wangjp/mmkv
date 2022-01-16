package mmdb

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/beego/beego/v2/core/logs"
)

//添加多个用户
func (db *MemoryDb) MmAddUsers(users map[string]string) {
	for u, p := range users {
		db.MmAddUser(u, p)
	}
}

//添加用户
func (db *MemoryDb) MmAddUser(username, password string) {
	if !strings.HasPrefix(username, _UserPrefix) { //用户名以"users."为前缀
		username = _UserPrefix + username
	}
	if len(password) != 32 { //密码必须是MD5后的值
		sum := md5.Sum([]byte(password))
		password = hex.EncodeToString(sum[:])
	}
	db.Store(username, password)
}

//校验用户
func (db *MemoryDb) MmCheckUser(username, password string) (bool, error) {
	name := username
	if !strings.HasPrefix(username, _UserPrefix) { //用户名以"users."为前缀
		username = _UserPrefix + username
	}
	if len(password) != 32 { //密码必须是MD5后的值
		sum := md5.Sum([]byte(password))
		password = hex.EncodeToString(sum[:])
	}
	if pswd, ok := db.Load(username); ok { //用户名存在
		if password == pswd.(string) { //密码匹配
			return true, nil
		} else { //密码不匹配
			logs.Debug(i18n("log_debug_user_pswd_err"), name)
			return false, fmt.Errorf("user_passsword_error")
		}
	} else { //用户名不存在
		logs.Debug(i18n("log_debug_user_not_exist"), name)
		return false, fmt.Errorf("user_noname")
	}
}

//读取单个标签
func (db *MemoryDb) MmReadSingle(key string) (val interface{}, ok bool) {
	logs.Debug(i18n("log_debug_read_single"), key)
	return db.Load(key)
}

//读取多个标签
func (db *MemoryDb) MmReadMulti(keys []string) (datas []RespMsg) {
	logs.Debug(i18n("log_debug_read_multi"), keys)
	for _, k := range keys {
		var msg RespMsg
		msg.Msg = k
		msg.Data, msg.Ok = db.Load(k)
		if !msg.Ok {
			msg.Data = ""
		}
		datas = append(datas, msg)
	}
	return
}

//写单个标签
func (db *MemoryDb) MmWriteSingle(key string, value interface{}) {
	logs.Debug(i18n("log_debug_write_single"), key)
	db.Store(key, value)
}

//写多个标签
func (db *MemoryDb) MmWriteMulti(maps map[string]interface{}) {
	logs.Debug(i18n("log_debug_write_multi"), maps)
	for k, v := range maps {
		db.Store(k, v)
	}
}

//删除单个标签
func (db *MemoryDb) MmDeleteSingle(key string) {
	logs.Debug(i18n("log_debug_delete_single"), key)
	db.Delete(key)
}

//删除多个标签
func (db *MemoryDb) MmDeleteMulti(keys []string) {
	logs.Debug(i18n("log_debug_delete_multi"), keys)
	for _, k := range keys {
		db.Delete(k)
	}
}

//通过错误码获取错误信息
func (db *MemoryDb) I18n(code string) (errstr string) {
	ecode, ok := textDictionary[code]
	if !ok {
		ecode = unDefined
	}
	msg, ok := ecode[db.Langtype]
	if !ok {
		msg = fmt.Sprintf("Undefined languige type:%s", db.Langtype)
	}
	return msg
}
