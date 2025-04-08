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
package tls

import (
	"context"
	"fmt"
	"net"

	"github.com/aberstone/tls_mitm_server/transport/proxy_connector"
)

type ProxyTLSDialer struct {
	*BaseTLSDialer
	connector proxy_connector.ProxyConnector
}

func (d *ProxyTLSDialer) DialTLS(ctx context.Context, network, addr string) (net.Conn, error) {
	d.opts.logger.Info(fmt.Sprintf("[TLS] 通过代理连接到 %s", addr))

	proxyConn, err := d.connector.Connect(ctx, d.opts.upstreamProxy, addr)
	if err != nil {
		d.opts.logger.Error(fmt.Sprintf("代理连接到 %s 失败", addr), err)
		return nil, err
	}

	return d.handshakeTLS(ctx, proxyConn, extractServerName(addr))
}
