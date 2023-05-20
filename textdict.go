package mmkv

var (
	textDictionary = map[string]map[string]string{ //文本字典
		//用于通讯的错误信息
		"user_unlogin": {
			"en-US": "You must log in first",
			"zh-CN": "需要先登录"},
		"user_noname": {
			"en-US": "User name [%s] not exist",
			"zh-CN": "用户名[%s]不存在"},
		"user_exist": {
			"en-US": "Add user failed:User name [%s] has exist",
			"zh-CN": "添加用户失败:用户名[%s]已存在"},
		"user_passsword_error": {
			"en-US": "Password error",
			"zh-CN": "密码错误"},
		"fcode_undefined": {
			"en-US": "Function code not defined:%v",
			"zh-CN": "功能码未定义:%v"},
		"comm_less_4byte": {
			"en-US": "The length of communication message shall be at least 4 bytes",
			"zh-CN": "通讯报文长度至少4字节"},
		"write_type_mismatch": {
			"en-US": "The new data type does not match the original data type, original type is: %v, new data type is: %v",
			"zh-CN": "写标签时新值与原有值类型不匹配,原有数据的类型为:%v,新数据类型为:%v"},
		"write_no_tag": {
			"en-US": "No valid label is set when writing label",
			"zh-CN": "写标签时未设定有效的标签"},
		"undefined_key": {
			"en-US": "Undefined key",
			"zh-CN": "未定义的标签"},
		"data_type_fail_add": {
			"en-US": "The data type is incorrect. Addition and subtraction operations cannot be performed. The type is:%v",
			"zh-CN": "数据类型不正确,不可执行加减运算. 数据类型是:%v"},
		"data_assert_fail": {
			"en-US": "Data type assertion failed, actual type is:%v",
			"zh-CN": "数据类型断言失败,实际类型为:%v"},
		"data_type_fail": {
			"en-US": "The data type is incorrect. The type is:%v",
			"zh-CN": "数据类型不正确, 数据类型是:%v"},
		"no_data_in_pipe": {
			"en-US": "There is no data in the pipeline",
			"zh-CN": "管道中已经没有数据"},
		"sys_reserved_key": {
			"en-US": "[%s] is system reserved keyword",
			"zh-CN": "[%s] 是系统保留关键字"},
		//用于日志的信息
		//错误日志
		"log_err_listen_faile": {
			"en-US": "Failed to start communication port listening:%s",
			"zh-CN": "启动通讯端口监听失败:%s"},
		"log_err_new_conn": {
			"en-US": "An error occurred while accessing a new connection:",
			"zh-CN": "接入新的连接时发生错误:"},
		"log_err_len_less_2byte": {
			"en-US": "the data length should not be less than 2 bytes",
			"zh-CN": "数据长度不能小于2字节"},
		"log_err_reading_conn": {
			"en-US": "An error occurred while reading the connection",
			"zh-CN": "读取连接时发生错误:"},
		"log_err_check_code": {
			"en-US": "Check crc16 code error",
			"zh-CN": "CRC16校验码错误"},
		"log_err_splite_data": {
			"en-US": "Split data error:%s",
			"zh-CN": "分割数据错误:%s"},
		"log_err_socket_write": {
			"en-US": "Socket connection write data error",
			"zh-CN": "Socket连接写数据错误"},
		"log_err_encode_json": {
			"en-US": "An error occurred while converting the result to JSON:",
			"zh-CN": "结果转换为json时发生错误:"},
		//Info日志
		"log_info_listen_port": {
			"en-US": "Start running and listens on the port:",
			"zh-CN": "内存数据库开始运行,监听端口:"},
		"log_info_new_conn": {
			"en-US": "New client connection:%s",
			"zh-CN": "新的客户端连接:%s"},
		"log_info_client_shutdown": {
			"en-US": "Client [%s] close",
			"zh-CN": "客户端连接 [%s] 关闭"},
		"log_info_user_login": {
			"en-US": "User [%s] login from client:%v",
			"zh-CN": "用户 [%s] 从客户端 [%v] 登录"},
		//调试日志
		"log_debug_terminator_received": {
			"en-US": "Terminator received",
			"zh-CN": "接收到结束符"},
		"log_debug_user_pswd_err": {
			"en-US": "User [%s] login faile, passwords do not match",
			"zh-CN": "用户 [%s] 登录失败, 密码不匹配"},
		"log_debug_read_single": {
			"en-US": "Read single key: %s",
			"zh-CN": "读取标签: %s"},
		"log_debug_read_multi": {
			"en-US": "Read Multi keys: %v",
			"zh-CN": "读取多个标签: %v"},
		"log_debug_write_single": {
			"en-US": "Write single key: {%s:%v}",
			"zh-CN": "写标签: {%s:%v}"},
		"log_debug_write_multi": {
			"en-US": "Write Multi keys: %v",
			"zh-CN": "批量写标签: %v"},
		"log_debug_delete_single": {
			"en-US": "Delete single key: %s",
			"zh-CN": "删除标签: %s"},
		"log_debug_delete_multi": {
			"en-US": "Delete Multi keys: %v",
			"zh-CN": "批量删除标签: %v"},
		"log_debug_self_increase": {
			"en-US": "Self increase: %v",
			"zh-CN": "标签自增: %v"},
		"log_debug_pipe_push": {
			"en-US": "Pipe push:{%s:%v}",
			"zh-CN": "压入管道:{%s:%v}"},
		"log_debug_pipe_pull": {
			"en-US": "Pipe pull:%s",
			"zh-CN": "从管道拉取数据:%s"},
		"log_debug_pipe_len": {
			"en-US": "Get pipe length:%s",
			"zh-CN": "获取管道的长度:%s"},
		"未定义": {
			"en-US": "Undifend",
			"zh-CN": "未定义"},
		"连接测试": {
			"en-US": "Ping",
			"zh-CN": "连接测试"},
		"用户登录": {
			"en-US": "Login",
			"zh-CN": "用户登录"},
		"写单个标签": {
			"en-US": "WriteSingleKey",
			"zh-CN": "写单个标签"},
		"写多个标签": {
			"en-US": "WriteMultiKey",
			"zh-CN": "写多个标签"},
		"读取单个标签": {
			"en-US": "ReadSingleKey",
			"zh-CN": "读取单个标签"},
		"读取多个标签": {
			"en-US": "ReadMultiKey",
			"zh-CN": "读取多个标签"},
		"删除单个标签": {
			"en-US": "DeleteSingleKey",
			"zh-CN": "删除单个标签"},
		"删除多个标签": {
			"en-US": "DeleteMultiKey",
			"zh-CN": "删除多个标签"},
		"数据自增": {
			"en-US": "SelfIncrease",
			"zh-CN": "数据自增"},
		"数据自减": {
			"en-US": "SelfDecrease",
			"zh-CN": "数据自减"},
		"压入管道": {
			"en-US": "PipePush",
			"zh-CN": "压入管道"},
		"从管道拉取(FIFO)": {
			"en-US": "PipeFiFoPull",
			"zh-CN": "从管道拉取(FIFO)"},
		"从管道拉取(FILO)": {
			"en-US": "PipeFiLoPull",
			"zh-CN": "从管道拉取(FILO)"},
		"获取管道当前长度": {
			"en-US": "PipeLenght",
			"zh-CN": "获取管道当前长度"},
		"读取管道": {
			"en-US": "PipeAll",
			"zh-CN": "读取管道"},
		"读取键": {
			"en-US": "GetKeys",
			"zh-CN": "读取键"},
		"读取用户列表": {
			"en-US": "GetUsers",
			"zh-CN": "读取用户列表"},
		"功能码": {
			"en-US": "Function Code",
			"zh-CN": "功能码"},
		"保留字节": {
			"en-US": "Reserved Bytes",
			"zh-CN": "保留字节"},
		"数据内容": {"en-US": "Data",
			"zh-CN": "数据内容"},
		"响应信息": {"en-US": "Respond Message",
			"zh-CN": "响应信息"},
		"内存数据库启动": {"en-US": "mmkv start",
			"zh-CN": "内存数据库启动"},
		"启动参数": {"en-US": "Start parameters",
			"zh-CN": "启动参数"},
		"注册用户": {"en-US": "Register an account",
			"zh-CN": "注册用户"},
	}
)
