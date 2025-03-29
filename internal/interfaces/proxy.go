package interfaces

import (
	"context"
	"net"
	"net/url"
)

// ProxyServer 定义代理服务器的基本接口
type ProxyServer interface {
	// Start 启动代理服务器
	Start(ctx context.Context) error
	// Stop 停止代理服务器
	Stop(ctx context.Context) error
	// GetProxyURL 获取代理URL
	GetProxyURL() *url.URL
}

// ConnectionHandler 定义连接处理器接口
type ConnectionHandler interface {
	// HandleConnection 处理新的连接
	HandleConnection(ctx context.Context, conn net.Conn) error
}

// TLSDialer 定义TLS连接创建接口
type TLSDialer interface {
	// DialTLS 建立TLS连接
	DialTLS(ctx context.Context, network, addr string) (net.Conn, error)
}

// CertificateGenerator 定义证书生成器接口
type CertificateGenerator interface {
	// GenerateCA 生成CA证书和私钥
	GenerateCA(certPath, keyPath string) error
	// GenerateCert 生成服务器证书
	GenerateCert(domain string) error
}
