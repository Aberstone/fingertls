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
	"net"
	"net/url"
	"time"

	utls "github.com/refraction-networking/utls"
)

// ProxyScheme 定义支持的代理协议类型
type ProxyScheme string

const (
	ProxySchemeHTTP   ProxyScheme = "http"
	ProxySchemeSocks5 ProxyScheme = "socks5"
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
