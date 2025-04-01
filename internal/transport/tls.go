/*
 * Copyright (C) 2024 aberstone
 *
 * This library is free software; you can redistribute it and/or
 * modify it under the terms of the GNU Lesser General Public
 * License as published by the Free Software Foundation; either
 * version 3 of the License, or (at your option) any later version.
 *
 * This library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public
 * License along with this library; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 */
package transport

import (
	"context"
	"fmt"
	"net"

	"github.com/aberstone/tls_mitm_server/internal/interfaces"
	"github.com/aberstone/tls_mitm_server/internal/logging"

	utls "github.com/refraction-networking/utls"
)

// BaseTLSDialer 提供基础TLS功能
type BaseTLSDialer struct {
	config DialerConfig
	logger *logging.Logger
}

func (d *BaseTLSDialer) handshakeTLS(ctx context.Context, conn net.Conn, serverName string) (net.Conn, error) {
	d.logger.Info(fmt.Sprintf("[TLS] 开始与 %s 进行TLS握手", serverName))

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
	case "socks5":
		logger.Info(fmt.Sprintf("[TLS] 创建SOCKS5代理TLS拨号器 (代理: %s)", config.UpstreamProxy.Host))
		connector = newSOCKS5ProxyConnector(config.Timeout, logger)
	default:
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
