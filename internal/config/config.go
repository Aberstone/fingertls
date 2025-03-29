package config

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/spf13/viper"

	"tls_mitm_server/internal/errors"
)

// TLSConfig TLS配置
type TLSConfig struct {
	CACertPath string `mapstructure:"ca_cert" json:"ca_cert"`
	CAKeyPath  string `mapstructure:"ca_key" json:"ca_key"`
	// 证书生成配置
	CertConfig struct {
		Organization string `mapstructure:"organization" json:"organization"`
		Country      string `mapstructure:"country" json:"country"`
		CommonName   string `mapstructure:"common_name" json:"common_name"`
		ValidYears   int    `mapstructure:"valid_years" json:"valid_years"`
	} `mapstructure:"cert_config" json:"cert_config"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Port          int    `mapstructure:"port" json:"port"`
	UpstreamProxy string `mapstructure:"upstream_proxy" json:"upstream_proxy"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level   string `mapstructure:"level" json:"level"`
	Format  string `mapstructure:"format" json:"format"`
	Output  string `mapstructure:"output" json:"output"`
	Verbose bool   `mapstructure:"verbose" json:"verbose"`
}

// Config 总配置结构
type Config struct {
	TLS   TLSConfig   `mapstructure:"tls" json:"tls"`
	Proxy ProxyConfig `mapstructure:"proxy" json:"proxy"`
	Log   LogConfig   `mapstructure:"log" json:"log"`
}

// LoadConfig 加载配置
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// 设置默认值
	setDefaults(v)

	// 如果提供了配置文件路径，则从文件加载
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, errors.NewError(errors.ErrConfiguration, "读取配置文件失败", err)
		}
	}

	// 从环境变量加载
	v.AutomaticEnv()
	v.SetEnvPrefix("MITM")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, errors.NewError(errors.ErrConfiguration, "解析配置失败", err)
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// setDefaults 设置默认配置
func setDefaults(v *viper.Viper) {
	v.SetDefault("tls.cert_config.organization", "MITM Proxy")
	v.SetDefault("tls.cert_config.country", "CN")
	v.SetDefault("tls.cert_config.common_name", "MITM Root CA")
	v.SetDefault("tls.cert_config.valid_years", 10)

	v.SetDefault("proxy.port", 8080)

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "text")
	v.SetDefault("log.output", "stdout")
	v.SetDefault("log.verbose", false)
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	// 验证 TLS 配置
	if config.TLS.CACertPath == "" || config.TLS.CAKeyPath == "" {
		return errors.NewError(errors.ErrConfiguration, "必须提供CA证书和私钥路径", nil)
	}

	// 验证代理配置
	if config.Proxy.Port <= 0 || config.Proxy.Port > 65535 {
		return errors.NewError(errors.ErrConfiguration, "无效的代理端口", nil)
	}

	// 验证日志配置
	switch config.Log.Level {
	case "debug", "info", "warn", "error":
	// 有效的日志级别
	default:
		return errors.NewError(errors.ErrConfiguration, "无效的日志级别", nil)
	}

	switch config.Log.Format {
	case "text", "json":
	// 有效的日志格式
	default:
		return errors.NewError(errors.ErrConfiguration, "无效的日志格式", nil)
	}

	return nil
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.NewError(errors.ErrConfiguration, "序列化配置失败", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return errors.NewError(errors.ErrConfiguration, "写入配置文件失败", err)
	}

	return nil
}
