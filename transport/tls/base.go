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

	utls "github.com/refraction-networking/utls"
)

type BaseTLSDialer struct {
	opts *Options
}

func extractServerName(addr string) string {
	host, _, _ := net.SplitHostPort(addr)
	return host
}

func (d *BaseTLSDialer) handshakeTLS(ctx context.Context, conn net.Conn, serverName string) (net.Conn, error) {
	d.opts.logger.Info(fmt.Sprintf("[TLS] 开始与 %s 进行TLS握手", serverName))

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

	d.opts.logger.Info("[TLS] 应用ClientHello预设...")
	if err := uConn.ApplyPreset(d.opts.sf()); err != nil {
		d.opts.logger.Error("应用ClientHello预设失败", err)
		return nil, fmt.Errorf("应用ClientHello预设失败: %w", err)
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- uConn.Handshake()
	}()

	select {
	case err := <-errChan:
		if err != nil {
			d.opts.logger.Error("TLS握手失败", err)
			return nil, fmt.Errorf("TLS握手失败: %w", err)
		}
		state := uConn.ConnectionState()
		d.opts.logger.Info(fmt.Sprintf("[TLS] 握手成功 - 协议: %s, 密码套件: %d", state.NegotiatedProtocol, state.CipherSuite))
	case <-ctx.Done():
		d.opts.logger.Error("TLS握手超时或被取消", ctx.Err())
		return nil, ctx.Err()
	}

	return uConn, nil
}

func (d *BaseTLSDialer) DialTLS(ctx context.Context, network, addr string) (net.Conn, error) {
	d.opts.logger.Info(fmt.Sprintf("[TLS] 直接连接到 %s", addr))

	tcpConn, err := (&net.Dialer{Timeout: d.opts.timeout}).DialContext(ctx, network, addr)
	if err != nil {
		d.opts.logger.Error(fmt.Sprintf("TCP连接到 %s 失败", addr), err)
		return nil, err
	}

	return d.handshakeTLS(ctx, tcpConn, extractServerName(addr))
}
