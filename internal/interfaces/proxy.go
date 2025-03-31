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
