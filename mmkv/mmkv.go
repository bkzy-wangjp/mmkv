package main

import (
	"flag"

	"github.com/bkzy-wangjp/mmkv"
)

var (
	MMKV_HOST     = flag.String("h", "127.0.0.1", "mmkv内存数据库的主机地址")
	MMKV_PORT     = flag.Int("p", 9646, "mmkv内存数据库的端口号")
	MMKV_USER     = flag.String("u", "admin", "mmkv内存数据库的登录用户名")
	MMKV_PSWD     = flag.String("psw", "admin123", "mmkv内存数据库的登录用户密码")
	MMKV_LANG     = flag.String("lang", "zh-CN", "语言类型")
	MMKV_LOGL     = flag.Int("logl", 7, "日志等级")
	MMKV_LOGP     = flag.String("logp", "./log", "日志存储路径")
	MMKV_LOGS     = flag.Int("logs", 100000, "日志文件最大行数")
	MMKV_LOGD     = flag.Int("logd", 180, "日志保存天数")
	MMKV_SHOWPATH = flag.Bool("logsp", false, "日志显示路径")
)

func main() {
	flag.Parse()
	mmkv.Run(map[string]string{*MMKV_USER: *MMKV_PSWD},
		map[string]interface{}{
			"ip":          *MMKV_HOST,
			"port":        *MMKV_PORT,
			"languige":    *MMKV_LANG,
			"loglevel":    *MMKV_LOGL,
			"logpath":     *MMKV_LOGP,
			"logsize":     *MMKV_LOGS,
			"logsavedays": *MMKV_LOGD,
			"logshowpath": *MMKV_SHOWPATH,
		})
}
