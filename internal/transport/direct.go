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
)

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
