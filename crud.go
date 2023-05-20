package mmkv

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"reflect"
	"time"

	"github.com/beego/beego/v2/core/logs"
)

//初始化mmkv
func (db *MemoryKeyValueMap) init(users map[string]string, langtype string) {
	db.Langtype = langtype
	db.MmAddUsers(users)
	db.Store(_KeysDict, make(map[string]time.Time))
}

//添加多个用户
func (db *MemoryKeyValueMap) MmAddUsers(users map[string]string) {
	for u, p := range users {
		if err := db.MmAddUser(u, p); err != nil {
			logs.Error(err)
		}
	}
}

//添加用户
func (db *MemoryKeyValueMap) MmAddUser(username, password string) error {
	if len(password) != 32 { //密码必须是MD5后的值
		sum := md5.Sum([]byte(password))
		password = hex.EncodeToString(sum[:])
	}

	if usersmp, ok := db.Load(_UsersDict); ok { //用户字典已存在
		if usersdict, ok := usersmp.(map[string]UserDict); ok { //不能转换为用户字典结构
			if _, ok := usersdict[username]; ok { //用户名已存在
				return fmt.Errorf(i18n("user_exist"), username)
			} else { //用户名不存在
				logs.Info("[%s] %s", username, i18n("注册用户"))
				usersdict[username] = UserDict{Password: password}
				db.Store(_UsersDict, usersdict)
			}
		}
	} else {
		logs.Info("[%s] %s", username, i18n("注册用户"))
		usersdict := make(map[string]UserDict)
		usersdict[username] = UserDict{Password: password}
		db.Store(_UsersDict, usersdict)
	}
	return nil
}

//校验用户
func (db *MemoryKeyValueMap) MmCheckUser(username, password string) (bool, error) {
	if len(password) != 32 { //密码必须是MD5后的值
		sum := md5.Sum([]byte(password))
		password = hex.EncodeToString(sum[:])
	}
	if usersmp, ok := db.Load(_UsersDict); ok { //用户字典已存在
		if usersdict, ok := usersmp.(map[string]UserDict); ok { //不能转换为用户字典结构
			if user, ok := usersdict[username]; ok { //用户名已存在
				if user.Password == password {
					return true, nil
				} else { //密码不匹配
					logs.Info(i18n("log_debug_user_pswd_err"), username)
					return false, fmt.Errorf(i18n("user_passsword_error"))
				}
			} else { //用户名不存在
				logs.Info(i18n("user_noname"), username)
				return false, fmt.Errorf(i18n("user_noname"), username)
			}
		}
	}
	usersdict := make(map[string]UserDict)
	db.Store(_UsersDict, usersdict)
	return false, fmt.Errorf(i18n("user_noname"), username)
}

//读取单个标签
func (db *MemoryKeyValueMap) MmReadSingle(key string) (val interface{}, ok bool) {
	//logs.Debug(i18n("log_debug_read_single"), key)
	return db.Load(key)
}

//读取多个标签
func (db *MemoryKeyValueMap) MmReadMulti(keys []string) (datas []RespMsg) {
	//logs.Debug(i18n("log_debug_read_multi"), keys)
	for _, k := range keys {
		var msg RespMsg
		msg.Msg = k
		msg.Data, msg.Ok = db.Load(k)
		if !msg.Ok {
			msg.Data = fmt.Sprintf(i18n("undefined_key"), k)
		}
		datas = append(datas, msg)
	}
	return
}

//读取用户名字典
func (db *MemoryKeyValueMap) MmGetUsersDict() []string {
	if usersmp, ok := db.Load(_UsersDict); ok {
		if usersdict, ok := usersmp.(map[string]UserDict); ok {
			var users []string
			for user := range usersdict {
				users = append(users, user)
			}
			return users
		}
	}
	return nil
}

/*
//读取用户键字典
func (db *MemoryKeyValueMap) MmGetKeysDictx() map[string]string {
	if keysmp, ok := db.Load(_KeysDict); ok {
		if dict, ok := keysmp.(map[string]time.Time); ok {
			keys := make(map[string]string)
			for k, v := range dict {
				keys[k] = v.Local().Format("2006-01-02T15:04:05.000")
			}
			return keys
		}
	}
	db.Store(_KeysDict, make(map[string]time.Time))
	return nil
}

//添加键到字典中
func (db *MemoryKeyValueMap) addToKeysDict(keys ...string) {
	if keysmp, ok := db.Load(_KeysDict); ok {
		if dict, ok := keysmp.(map[string]time.Time); ok {
			for _, k := range keys {
				dict[k] = time.Now()
			}
			db.Store(_KeysDict, dict)
		}
	} else {
		db.Store(_KeysDict, make(map[string]time.Time))
	}
}

//从字典中删除键
func (db *MemoryKeyValueMap) deleteFromKeysDict(keys ...string) {
	if keysmp, ok := db.Load(_KeysDict); ok {
		if dict, ok := keysmp.(map[string]time.Time); ok {
			for _, k := range keys {
				delete(dict, k)
			}
			db.Store(_KeysDict, dict)
		}
	} else {
		db.Store(_KeysDict, make(map[string]time.Time))
	}
}
*/
//写单个标签
func (db *MemoryKeyValueMap) MmWriteSingle(key string, value interface{}) (interface{}, error) {
	//logs.Debug(i18n("log_debug_write_single"), key, value)
	if isSysReservedKey(key) { //检查是否关键字
		return value, fmt.Errorf(i18n("sys_reserved_key"), key)
	}

	oldv, ok := db.Load(key) //加载标签
	if ok {                  //标签存在
		oldisnum, oldtype := isNumber(oldv)
		newisnum, newtype := isNumber(value)
		if oldtype != nil {
			if oldtype != newtype { //数据类型不同
				if !oldisnum || !newisnum { //且至少有一个不是数字
					return oldv, fmt.Errorf(i18n("write_type_mismatch"), oldtype, newtype) //返回错误信息
				}
			}
		}
	} //else {
	//	db.addToKeysDict(key) //添加新键
	//}
	db.Store(key, value)
	return value, nil
}

//写多个标签
//返回新创建的标签数
func (db *MemoryKeyValueMap) MmWriteMulti(maps map[string]interface{}) []RespMsg {
	var resps []RespMsg
	for k, v := range maps {
		var resp RespMsg
		creat, err := db.MmWriteSingle(k, v)
		if err != nil {
			resp = MakeRespMsg(false, k, err.Error())
		} else {
			resp = MakeRespMsg(true, k, creat)
		}
		resps = append(resps, resp)
	}
	return resps
}

//删除单个标签
//如果删除成功(标签存在),返回1
//如果标签不存在,返回0
func (db *MemoryKeyValueMap) MmDeleteSingle(key string) int64 {
	//logs.Debug(i18n("log_debug_delete_single"), key)
	_, ok := db.Load(key)
	db.Delete(key)
	//删除keys字典中的记录
	//db.deleteFromKeysDict(key)
	if ok {
		return 0
	} else {
		return 1
	}
}

//删除多个标签
//返回删除成功的标签数
func (db *MemoryKeyValueMap) MmDeleteMulti(keys []string) int64 {
	//logs.Debug(i18n("log_debug_delete_multi"), keys)
	var deleted int64 = 0
	for _, k := range keys {
		_, ok := db.Load(k)
		if !ok {
			deleted += 1
		}
		db.Delete(k)
	}
	//db.deleteFromKeysDict(keys...)
	return deleted
}

//通过错误码获取错误信息
/*
func (db *MemoryKeyValueMap) getErrorMsg(code string) (errstr string) {
	ecode, ok := textDictionary[code]
	if !ok {
		return code
	}
	msg, ok := ecode[db.Langtype]
	if !ok {
		msg = code
	}
	return msg
}
*/

//标签自增
func (db *MemoryKeyValueMap) MmSelfIncrease(key string, value int64) (interface{}, error) {
	if isSysReservedKey(key) { //检查是否关键字
		return value, fmt.Errorf(i18n("sys_reserved_key"), key)
	}

	//logs.Debug(i18n("log_debug_self_increase"), key)
	oldv, ok := db.Load(key) //加载标签
	if ok {                  //标签存在
		newv, err := selfAdd(oldv, value) //值自加
		if err != nil {                   //数据类型不可加减运算
			return 0, err
		} else { //自加完成
			db.Store(key, newv) //保存
			return newv, nil    //返回新值
		}
	} else { //标签不存在
		var init int64 = 1
		db.Store(key, init) //新建标签
		//db.addToKeysDict(key) //添加新键
		return init, nil //返回新值
	}
}

//压入管道
//返回管道长度和错误信息
func (db *MemoryKeyValueMap) MmPipePush(key string, value interface{}) (int, error) {
	if isSysReservedKey(key) { //检查是否关键字
		return 0, fmt.Errorf(i18n("sys_reserved_key"), key)
	}

	//logs.Debug(i18n("log_debug_pipe_push"), key, value)
	oldv, ok := db.Load(key) //加载标签
	if ok {                  //标签存在
		oldtype := reflect.TypeOf(oldv)
		newtype := reflect.TypeOf(value)
		if oldtype.Kind() != reflect.Slice { //原值非切片类型
			if oldtype != newtype { //数据类型不同
				return 0, fmt.Errorf(i18n("write_type_mismatch"), oldtype, newtype) //返回错误信息
			} else {
				var vslice []interface{}
				vslice = append(vslice, oldv, value)
				db.Store(key, vslice)
				return len(vslice), nil
			}
		} else { //原值为切片类型
			vslice, ok := oldv.([]interface{})
			if !ok {
				return 0, fmt.Errorf(i18n("data_assert_fail"), oldtype) //返回错误信息
			} else {
				vslice = append(vslice, value)
				db.Store(key, vslice)
				return len(vslice), nil
			}
		}
	} else { //标签不存在
		var vslice []interface{}
		vslice = append(vslice, value)
		db.Store(key, vslice)
		//db.addToKeysDict(key) //添加新键
		return len(vslice), nil
	}
}

//从管道拉取数据
//返回管道剩余长度、数据、错误信息
func (db *MemoryKeyValueMap) MmPipePull(fc byte, key string) (int, interface{}, error) {
	if isSysReservedKey(key) { //检查是否关键字
		return 0, 0, fmt.Errorf(i18n("sys_reserved_key"), key)
	}

	//logs.Debug(i18n("log_debug_pipe_pull"), key)
	oldv, ok := db.Load(key) //加载标签
	if ok {                  //标签存在
		oldtype := reflect.TypeOf(oldv)
		if oldtype.Kind() != reflect.Slice { //原值非切片类型
			return 0, nil, fmt.Errorf(i18n("data_type_fail"), oldtype) //数据类型错误
		} else { //原值为切片类型
			vslice, ok := oldv.([]interface{})
			if !ok {
				return 0, nil, fmt.Errorf(i18n("data_assert_fail"), oldtype) //数据断言错误
			} else {
				if len(vslice) > 0 { //数据长度大于0
					var val interface{}
					if fc == _FC_PipeFiFoPull { //先入先出
						val = vslice[0]
						vslice = vslice[1:]
					} else { //先进后出
						val, vslice = vslice[len(vslice)-1], vslice[:len(vslice)-1]
					}
					db.Store(key, vslice)
					return len(vslice), val, nil
				} else { //切片中已经没有数据
					return -1, nil, fmt.Errorf(i18n("no_data_in_pipe")) //管道中已经没有数据
				}
			}
		}
	} else { //标签不存在
		return 0, nil, fmt.Errorf(i18n("undefined_key"), key) //未定义的标签
	}
}

//获取管道中的所有数据
//返回管道剩余长度和错误信息
func (db *MemoryKeyValueMap) MmPipeAll(key string) ([]interface{}, error) {
	//logs.Debug(i18n("log_debug_pipe_len"), key)
	oldv, ok := db.Load(key) //加载标签
	if ok {                  //标签存在
		oldtype := reflect.TypeOf(oldv)
		if oldtype.Kind() != reflect.Slice { //原值非切片类型
			return nil, fmt.Errorf(i18n("data_type_fail"), oldtype) //数据类型错误
		} else { //原值为切片类型
			vslice, ok := oldv.([]interface{})
			if !ok {
				return nil, fmt.Errorf(i18n("data_assert_fail"), oldtype) //数据断言错误
			} else {
				if len(vslice) > 0 { //数据长度大于0
					val := make([]interface{}, len(vslice))
					copy(val, vslice)
					vslice = vslice[0:0]
					db.Store(key, vslice)
					return val, nil
				} else { //切片中已经没有数据
					return nil, fmt.Errorf(i18n("no_data_in_pipe")) //管道中已经没有数据
				}
			}
		}
	} else { //标签不存在
		return nil, fmt.Errorf(i18n("undefined_key"), key) //未定义的标签
	}
}

//获取管道长度
//返回管道剩余长度和错误信息
func (db *MemoryKeyValueMap) MmPipeLength(key string) (int, error) {
	//logs.Debug(i18n("log_debug_pipe_len"), key)
	oldv, ok := db.Load(key) //加载标签
	if ok {                  //标签存在
		oldtype := reflect.TypeOf(oldv)
		if oldtype.Kind() != reflect.Slice { //原值非切片类型
			return 0, fmt.Errorf(i18n("data_type_fail"), oldtype) //数据类型错误
		} else { //原值为切片类型
			vslice, ok := oldv.([]interface{})
			if !ok {
				return 0, fmt.Errorf(i18n("data_assert_fail"), oldtype) //数据断言错误
			} else {
				return len(vslice), nil
			}
		}
	} else { //标签不存在
		return 0, fmt.Errorf(i18n("undefined_key"), key) //未定义的标签
	}
}
