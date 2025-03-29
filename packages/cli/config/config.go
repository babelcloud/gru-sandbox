package config

import (
	"github.com/spf13/viper"
)

// CommandConfig 命令配置结构
type CommandConfig struct {
	Implementation string                   `mapstructure:"implementation"`
	SubCommands    map[string]CommandConfig `mapstructure:"sub_commands,omitempty"`
}

// Config 全局配置结构
type Config struct {
	Commands map[string]CommandConfig `mapstructure:"commands"`
}

var globalConfig Config

// InitConfig 初始化配置
func InitConfig() error {
	viper.SetConfigName("gbox")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.gbox")

	// 设置默认值
	viper.SetDefault("commands", map[string]CommandConfig{
		"box": {
			Implementation: "bash",
			SubCommands: map[string]CommandConfig{
				"create":  {Implementation: "bash"},
				"delete":  {Implementation: "bash"},
				"list":    {Implementation: "bash"},
				"exec":    {Implementation: "bash"},
				"inspect": {Implementation: "bash"},
				"start":   {Implementation: "bash"},
				"stop":    {Implementation: "bash"},
				"reclaim": {Implementation: "bash"},
			},
		},
		"cluster": {
			Implementation: "bash",
			SubCommands: map[string]CommandConfig{
				"setup":   {Implementation: "bash"},
				"cleanup": {Implementation: "bash"},
			},
		},
		"mcp": {
			Implementation: "bash",
			SubCommands: map[string]CommandConfig{
				"export": {Implementation: "bash"},
			},
		},
	})

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	return viper.Unmarshal(&globalConfig)
}

// GetCommandConfig 获取命令配置
func GetCommandConfig(cmdName string, subCmdName string) CommandConfig {
	if config, ok := globalConfig.Commands[cmdName]; ok {
		// 如果有子命令，返回子命令配置
		if subCmdName != "" {
			if subConfig, ok := config.SubCommands[subCmdName]; ok {
				return subConfig
			}
		}
		// 如果没有子命令或子命令配置不存在，返回主命令配置
		return config
	}
	// 默认使用bash脚本
	return CommandConfig{Implementation: "bash"}
}

// IsBashScript 判断是否使用bash脚本
func (c CommandConfig) IsBashScript() bool {
	return c.Implementation == "bash"
}
