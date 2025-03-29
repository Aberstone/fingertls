package transport

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"tls_mitm_server/internal/interfaces"
	"tls_mitm_server/internal/logging"

	utls "github.com/refraction-networking/utls"
)

// ProxyScheme 定义支持的代理协议类型
type ProxyScheme string

const (
	ProxySchemeHTTP ProxyScheme = "http"

	// 未来可以添加更多协议支持
	// ProxySchemeSocks5 ProxyScheme = "socks5"
)

// DialerConfig TLS拨号器配置
type DialerConfig struct {
	// TLS配置
	GetHelloSpec func() *utls.ClientHelloSpec
	Timeout      time.Duration

	// 可选的上游代理配置
	UpstreamProxy *url.URL
}

// ProxyConnector 代理连接器接口
type ProxyConnector interface {
	// Connect 建立到目标地址的代理连接
	Connect(ctx context.Context, proxyURL *url.URL, targetAddr string) (net.Conn, error)
}

// BaseTLSDialer 提供基础TLS功能
type BaseTLSDialer struct {
	config DialerConfig
	logger *logging.Logger
}

// type debugConn struct {
// 	net.Conn
// }

// func (d *debugConn) Write(p []byte) (int, error) {
// 	fmt.Printf("[DEBUG] Write %d bytes: %x\n", len(p), p)
// 	return d.Conn.Write(p)
// }

// func (d *debugConn) Read(p []byte) (int, error) {
// 	n, err := d.Conn.Read(p)
// 	fmt.Printf("[DEBUG] Read %d bytes: %x\n", n, p[:n])
// 	return n, err
// }

func (d *BaseTLSDialer) handshakeTLS(ctx context.Context, conn net.Conn, serverName string) (net.Conn, error) {
	d.logger.Info(fmt.Sprintf("[TLS] 开始与 %s 进行TLS握手", serverName))

	// debug := &debugConn{conn}

	config := &utls.Config{
		ServerName:             serverName,
		NextProtos:             []string{"h2", "http/1.1"},
		InsecureSkipVerify:     true,
		SessionTicketsDisabled: true,
	}

	uConn := utls.UClient(conn, config, utls.ClientHelloID{
		Client:  "Custom",
		Version: "1.0",
		Seed:    nil,
	})

	d.logger.Info("[TLS] 应用ClientHello预设...")
	if err := uConn.ApplyPreset(d.config.GetHelloSpec()); err != nil {
		d.logger.Error("应用ClientHello预设失败", err)
		return nil, fmt.Errorf("应用ClientHello预设失败: %w", err)
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- uConn.Handshake()
	}()

	select {
	case err := <-errChan:
		if err != nil {
			d.logger.Error("TLS握手失败", err)
			return nil, fmt.Errorf("TLS握手失败: %w", err)
		}
		state := uConn.ConnectionState()
		d.logger.Info(fmt.Sprintf("[TLS] 握手成功 - 协议: %s, 密码套件: %d", state.NegotiatedProtocol, state.CipherSuite))
	case <-ctx.Done():
		d.logger.Error("TLS握手超时或被取消", ctx.Err())
		return nil, ctx.Err()
	}

	return uConn, nil
}

func (d *BaseTLSDialer) extractServerName(addr string) string {
	host, _, _ := net.SplitHostPort(addr)
	return host
}

// DirectTLSDialer 实现直接TLS连接
type DirectTLSDialer struct {
	*BaseTLSDialer
}

func (d *DirectTLSDialer) DialTLS(ctx context.Context, network, addr string) (net.Conn, error) {
	d.logger.Info(fmt.Sprintf("[TLS] 直接连接到 %s", addr))

	tcpConn, err := (&net.Dialer{Timeout: d.config.Timeout}).DialContext(ctx, network, addr)
	if err != nil {
		d.logger.Error(fmt.Sprintf("TCP连接到 %s 失败", addr), err)
		return nil, err
	}

	return d.handshakeTLS(ctx, tcpConn, d.extractServerName(addr))
}

// ProxyTLSDialer 实现通过代理的TLS连接
type ProxyTLSDialer struct {
	*BaseTLSDialer
	connector ProxyConnector
}

func (d *ProxyTLSDialer) DialTLS(ctx context.Context, network, addr string) (net.Conn, error) {
	d.logger.Info(fmt.Sprintf("[TLS] 通过代理连接到 %s", addr))

	proxyConn, err := d.connector.Connect(ctx, d.config.UpstreamProxy, addr)
	if err != nil {
		d.logger.Error(fmt.Sprintf("代理连接到 %s 失败", addr), err)
		return nil, err
	}

	return d.handshakeTLS(ctx, proxyConn, d.extractServerName(addr))
}

// HTTPProxyConnector 实现HTTP代理连接
type HTTPProxyConnector struct {
	timeout time.Duration
	logger  *logging.Logger
}

func newHTTPProxyConnector(timeout time.Duration, logger *logging.Logger) *HTTPProxyConnector {
	return &HTTPProxyConnector{
		timeout: timeout,
		logger:  logger,
	}
}

func (c *HTTPProxyConnector) Connect(ctx context.Context, proxyURL *url.URL, targetAddr string) (net.Conn, error) {
	c.logger.Info(fmt.Sprintf("[UPSTREAM] 连接到代理服务器 %s", proxyURL.Host))

	// 连接到代理服务器
	dialer := &net.Dialer{Timeout: c.timeout}
	conn, err := dialer.DialContext(ctx, "tcp", proxyURL.Host)
	if err != nil {
		c.logger.Error(fmt.Sprintf("连接代理服务器 %s 失败", proxyURL.Host), err)
		return nil, err
	}

	// 发送CONNECT请求
	if err := c.sendConnectRequest(conn, targetAddr, proxyURL); err != nil {
		conn.Close()
		c.logger.Error(fmt.Sprintf("发送CONNECT请求到 %s 失败", proxyURL.Host), err)
		return nil, err
	}

	c.logger.Info(fmt.Sprintf("[UPSTREAM] 成功建立到 %s 的隧道连接", targetAddr))
	return conn, nil
}

func (c *HTTPProxyConnector) sendConnectRequest(conn net.Conn, targetAddr string, proxyURL *url.URL) error {
	// 准备认证信息
	var auth string
	if proxyURL.User != nil {
		if password, ok := proxyURL.User.Password(); ok {
			auth = base64.StdEncoding.EncodeToString([]byte(
				fmt.Sprintf("%s:%s", proxyURL.User.Username(), password),
			))
			c.logger.Info("[UPSTREAM] 使用认证信息")
		}
	}

	// 发送CONNECT请求
	req := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", targetAddr, targetAddr)
	if auth != "" {
		req += "Proxy-Authorization: Basic " + auth + "\r\n"
	}
	req += "\r\n"

	c.logger.Info(fmt.Sprintf("[UPSTREAM] 发送CONNECT请求到 %s", targetAddr))

	if _, err := conn.Write([]byte(req)); err != nil {
		return err
	}

	// 读取响应
	br := bufio.NewReader(conn)
	resp, err := readHTTPResponse(br)
	if err != nil {
		return err
	}

	// 处理缓冲区中的剩余数据
	if br.Buffered() > 0 {
		buf := make([]byte, br.Buffered())
		br.Read(buf)
	}

	// 检查响应状态
	if !strings.Contains(resp, "200") {
		c.logger.Error(fmt.Sprintf("代理服务器返回非200状态: %s", resp), nil)
		return fmt.Errorf("代理连接失败: %s", resp)
	}

	return nil
}

// NewTLSDialer 创建新的TLS拨号器
func NewTLSDialer(config DialerConfig, logger *logging.Logger) interfaces.TLSDialer {
	if config.UpstreamProxy == nil {
		logger.Info("[TLS] 创建直连TLS拨号器")
		return &DirectTLSDialer{
			BaseTLSDialer: &BaseTLSDialer{
				config: config,
				logger: logger,
			},
		}
	}

	var connector ProxyConnector
	switch config.UpstreamProxy.Scheme {
	case "http", "https":
		logger.Info(fmt.Sprintf("[TLS] 创建HTTP代理TLS拨号器 (代理: %s)", config.UpstreamProxy.Host))
		connector = newHTTPProxyConnector(config.Timeout, logger)
	// 可以在这里添加更多代理类型的支持
	// case "socks5":
	//
	//	connector = newSOCKS5ProxyConnector(config.Timeout, logger)
	default:
		// 默认使用HTTP代理
		logger.Info(fmt.Sprintf("[TLS] 使用默认HTTP代理TLS拨号器 (代理: %s)", config.UpstreamProxy.Host))
		connector = newHTTPProxyConnector(config.Timeout, logger)
	}

	return &ProxyTLSDialer{
		BaseTLSDialer: &BaseTLSDialer{
			config: config,
			logger: logger,
		},
		connector: connector,
	}
}

func readHTTPResponse(r *bufio.Reader) (string, error) {
	var sb strings.Builder
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return "", err
		}
		sb.WriteString(line)
		if line == "\r\n" {
			break
		}
	}
	return sb.String(), nil
}
