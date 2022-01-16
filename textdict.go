package mmdb

var (
	textDictionary = map[string]map[string]string{ //文本字典
		//用于通讯的错误信息
		"user_unlogin": {
			"en-US": "You must log in first",
			"zh-CN": "需要先登录"},
		"user_noname": {
			"en-US": "User name not exist",
			"zh-CN": "用户名不存在"},
		"user_passsword_error": {
			"en-US": "Password error",
			"zh-CN": "密码错误"},
		"fcode_undefined": {
			"en-US": "Function code not defined:%s",
			"zh-CN": "功能码未定义:%s"},
		"comm_less_4byte": {
			"en-US": "The length of communication message shall be at least 4 bytes",
			"zh-CN": "通讯报文长度至少4字节"},
		//用于日志的信息
		//错误日志
		"log_err_listen_faile": {
			"en-US": "[MicMmdb] Failed to start communication port listening:%s",
			"zh-CN": "[MicMmdb] 启动通讯端口监听失败:%s"},
		"log_err_new_conn": {
			"en-US": "[MicMmdb] An error occurred while accessing a new connection:",
			"zh-CN": "[MicMmdb] 接入新的连接时发生错误:"},
		"log_err_len_less_2byte": {
			"en-US": "the data length should not be less than 2 bytes",
			"zh-CN": "数据长度不能小于2字节"},
		"log_err_reading_conn": {
			"en-US": "[MicMmdb] An error occurred while reading the connection",
			"zh-CN": "[MicMmdb] 读取连接时发生错误:"},
		"log_err_check_code": {
			"en-US": "[MicMmdb] Check crc16 code error",
			"zh-CN": "[MicMmdb] CRC16校验码错误"},
		"log_err_splite_data": {
			"en-US": "[MicMmdb] Split data error:%s",
			"zh-CN": "[MicMmdb] 分割数据错误:%s"},
		"log_err_socket_write": {
			"en-US": "[MicMmdb] Socket connection write data error",
			"zh-CN": "[MicMmdb] Socket连接写数据错误"},
		"log_err_encode_json": {
			"en-US": "[MicMmdb] An error occurred while converting the result to JSON:",
			"zh-CN": "[MicMmdb] 结果转换为json时发生错误:"},
		//Info日志
		"log_info_listen_port": {
			"en-US": "[MicMmdb] Start running and listens on the port:",
			"zh-CN": "[MicMmdb] 内存数据库开始运行,监听端口:"},
		"log_info_new_conn": {
			"en-US": "[MicMmdb] New client connection:%s",
			"zh-CN": "[MicMmdb] 新的客户端连接:%s"},
		"log_info_client_shutdown": {
			"en-US": "[MicMmdb] Client [%s] close",
			"zh-CN": "[MicMmdb] 客户端连接 [%s] 关闭"},
		"log_info_user_login": {
			"en-US": "[MicMmdb] User [%s] login from client:%v",
			"zh-CN": "[MicMmdb] 用户 [%s] 从客户端 [%v] 登录"},
		//调试日志
		"log_debug_terminator_received": {
			"en-US": "[MicMmdb] Terminator received",
			"zh-CN": "[MicMmdb] 接收到结束符"},
		"log_debug_user_pswd_err": {
			"en-US": "[MicMmdb] User [%s] login faile, passwords do not match",
			"zh-CN": "[MicMmdb] 用户 [%s] 登录失败, 密码不匹配"},
		"log_debug_read_single": {
			"en-US": "[MicMmdb] Read single key: %s",
			"zh-CN": "[MicMmdb] 读取标签: %s"},
		"log_debug_read_multi": {
			"en-US": "[MicMmdb] Read Multi keys: %v",
			"zh-CN": "[MicMmdb] 读取多个标签: %v"},
		"log_debug_write_single": {
			"en-US": "[MicMmdb] Write single key: %s",
			"zh-CN": "[MicMmdb] 写标签: %s"},
		"log_debug_write_multi": {
			"en-US": "[MicMmdb] Write Multi keys: %v",
			"zh-CN": "[MicMmdb] 批量写标签: %v"},
		"log_debug_delete_single": {
			"en-US": "[MicMmdb] Delete single key: %s",
			"zh-CN": "[MicMmdb] 删除标签: %s"},
		"log_debug_delete_multi": {
			"en-US": "[MicMmdb] Delete Multi keys: %v",
			"zh-CN": "[MicMmdb] 批量删除标签: %v"},
	}
	//未定义文本
	unDefined = map[string]string{"zh-CN": "未定义的文本字典", "en-US": "Undefined code"}
)
