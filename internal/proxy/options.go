package proxy

import (
	"net/url"

	"tls_mitm_server/internal/interfaces"
	"tls_mitm_server/internal/logging"
)

// Options 代理服务器选项
type Options struct {
	// 基本配置
	Port          int
	UpstreamProxy *url.URL
	TLSDialer     interfaces.TLSDialer

	// CA证书配置
	CACert     []byte
	CAKey      []byte
	CertConfig *CertConfig

	// 功能组件
	Logger *logging.Logger

	// 功能开关
	Verbose bool
}

// CertConfig 证书配置
type CertConfig struct {
	Organization string
	Country      string
	CommonName   string
	ValidYears   int
}

// Option 代理选项设置函数
type Option func(*Options)

// WithPort 设置端口
func WithPort(port int) Option {
	return func(opts *Options) {
		opts.Port = port
	}
}

// WithUpstreamProxy 设置上游代理
func WithUpstreamProxy(upstreamURL *url.URL) Option {
	return func(opts *Options) {
		opts.UpstreamProxy = upstreamURL
	}
}

// WithTLSDialer 设置TLS连接器
func WithTLSDialer(dialer interfaces.TLSDialer) Option {
	return func(opts *Options) {
		opts.TLSDialer = dialer
	}
}

// WithCACert 设置CA证书
func WithCACert(cert, key []byte) Option {
	return func(opts *Options) {
		opts.CACert = cert
		opts.CAKey = key
	}
}

// WithCertConfig 设置证书配置
func WithCertConfig(config *CertConfig) Option {
	return func(opts *Options) {
		opts.CertConfig = config
	}
}

// WithLogger 设置日志记录器
func WithLogger(logger *logging.Logger) Option {
	return func(opts *Options) {
		opts.Logger = logger
	}
}

// WithVerbose 设置详细日志
func WithVerbose(verbose bool) Option {
	return func(opts *Options) {
		opts.Verbose = verbose
	}
}

// DefaultOptions 返回默认选项
func DefaultOptions() *Options {
	return &Options{
		Port:    8080,
		Verbose: false,
	}
}
