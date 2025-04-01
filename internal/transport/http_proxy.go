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
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/aberstone/tls_mitm_server/internal/logging"
)

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
