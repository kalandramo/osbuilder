package util

import (
	"strings"

	"github.com/spf13/viper"
)

// OnInitialize 复刻 onexstack/pkg/core/config.go: OnInitialize。
// 设置配置文件名、搜索目录、环境变量前缀，并在命令执行前把内容读入 viper。
//
// 注意：configFile 改为 func() *string（延迟求值），而不是构造期已求值的 *string。
// 原因：构造根命令时命令行 flag 尚未解析，若构造期就 viper.GetString("config") 会得到 ""，
// 导致 --config 永远失效。改为在 OnInitialize 闭包真正执行时（命令执行前、flag 已解析）
// 再动态取 --config 的值。
func OnInitialize(configFileFn func() *string, envPrefix string, loadDirs []string, defaultConfigName string) func() {
	return func() {
		if configFile := configFileFn(); configFile != nil {
			// 从命令行 --config 指定的配置文件中读取
			viper.SetConfigFile(*configFile)
		} else {
			for _, dir := range loadDirs {
				viper.AddConfigPath(dir)
			}
			viper.SetConfigType("yaml")
			viper.SetConfigName(defaultConfigName)
		}

		// 读取匹配的环境变量
		viper.AutomaticEnv()
		viper.SetEnvPrefix(envPrefix)
		// 将 key 中的 '.' 和 '-' 替换为 '_'，如 user.name -> OSCTL_USER_NAME
		replacer := strings.NewReplacer(".", "_", "-", "_")
		viper.SetEnvKeyReplacer(replacer)

		// 读取配置文件（找不到也不报错，使用默认值）
		_ = viper.ReadInConfig()
	}
}
