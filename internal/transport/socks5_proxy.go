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
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/url"
	"time"

	"github.com/aberstone/tls_mitm_server/internal/errors"
	"github.com/aberstone/tls_mitm_server/internal/logging"
)

const (
	socks5Version = 0x05

	// 认证方法
	authNone         = 0x00
	authPassword     = 0x02
	authNoAcceptable = 0xFF

	// 命令类型
	cmdConnect = 0x01

	// 地址类型
	addrTypeIPv4   = 0x01
	addrTypeDomain = 0x03
	addrTypeIPv6   = 0x04

	// 响应状态
	respSucceeded = 0x00
	respError     = 0x01
)

// SOCKS5ProxyConnector 实现SOCKS5代理连接
type SOCKS5ProxyConnector struct {
	timeout time.Duration
	logger  *logging.Logger
}

func newSOCKS5ProxyConnector(timeout time.Duration, logger *logging.Logger) *SOCKS5ProxyConnector {
	return &SOCKS5ProxyConnector{
		timeout: timeout,
		logger:  logger,
	}
}

func (c *SOCKS5ProxyConnector) Connect(ctx context.Context, proxyURL *url.URL, targetAddr string) (net.Conn, error) {
	c.logger.Info(fmt.Sprintf("[SOCKS5] 连接到代理服务器 %s", proxyURL.Host))

	// 连接到代理服务器
	dialer := &net.Dialer{Timeout: c.timeout}
	conn, err := dialer.DialContext(ctx, "tcp", proxyURL.Host)
	if err != nil {
		c.logger.Error(fmt.Sprintf("连接SOCKS5代理服务器 %s 失败", proxyURL.Host), err)
		return nil, errors.NewError(errors.ErrProxy, "连接SOCKS5代理服务器失败", err)
	}

	// 进行握手
	if err := c.handshake(conn, proxyURL); err != nil {
		conn.Close()
		return nil, err
	}

	// 发送连接请求
	if err := c.connectTarget(conn, targetAddr); err != nil {
		conn.Close()
		return nil, err
	}

	c.logger.Info(fmt.Sprintf("[SOCKS5] 成功建立到 %s 的连接", targetAddr))
	return conn, nil
}

func (c *SOCKS5ProxyConnector) handshake(conn net.Conn, proxyURL *url.URL) error {
	c.logger.Info("[SOCKS5] 开始握手...")

	// 发送版本和支持的认证方法
	var methods []byte
	if proxyURL.User != nil {
		methods = []byte{authNone, authPassword}
	} else {
		methods = []byte{authNone}
	}

	request := make([]byte, 2+len(methods))
	request[0] = socks5Version
	request[1] = byte(len(methods))
	copy(request[2:], methods)

	if _, err := conn.Write(request); err != nil {
		c.logger.Error("发送SOCKS5握手请求失败", err)
		return errors.NewError(errors.ErrProxy, "发送SOCKS5握手请求失败", err)
	}

	// 读取服务器响应
	response := make([]byte, 2)
	if _, err := io.ReadFull(conn, response); err != nil {
		c.logger.Error("读取SOCKS5握手响应失败", err)
		return errors.NewError(errors.ErrProxy, "读取SOCKS5握手响应失败", err)
	}

	if response[0] != socks5Version {
		return errors.NewError(errors.ErrProxy, fmt.Sprintf("不支持的SOCKS版本: %d", response[0]), nil)
	}

	// 处理认证
	switch response[1] {
	case authNone:
		c.logger.Info("[SOCKS5] 无需认证")
	case authPassword:
		c.logger.Info("[SOCKS5] 使用用户名密码认证")
		if err := c.authenticate(conn, proxyURL); err != nil {
			return err
		}
	case authNoAcceptable:
		return errors.NewError(errors.ErrProxy, "没有可接受的认证方法", nil)
	default:
		return errors.NewError(errors.ErrProxy, fmt.Sprintf("不支持的认证方法: %d", response[1]), nil)
	}

	return nil
}

func (c *SOCKS5ProxyConnector) authenticate(conn net.Conn, proxyURL *url.URL) error {
	if proxyURL.User == nil {
		return errors.NewError(errors.ErrProxy, "需要认证信息但未提供", nil)
	}

	username := proxyURL.User.Username()
	password, _ := proxyURL.User.Password()

	// 构造认证请求
	request := make([]byte, 3+len(username)+len(password))
	request[0] = 0x01 // 认证子版本
	request[1] = byte(len(username))
	copy(request[2:], username)
	request[2+len(username)] = byte(len(password))
	copy(request[3+len(username):], password)

	if _, err := conn.Write(request); err != nil {
		c.logger.Error("发送认证请求失败", err)
		return errors.NewError(errors.ErrProxy, "发送认证请求失败", err)
	}

	// 读取认证响应
	response := make([]byte, 2)
	if _, err := io.ReadFull(conn, response); err != nil {
		c.logger.Error("读取认证响应失败", err)
		return errors.NewError(errors.ErrProxy, "读取认证响应失败", err)
	}

	if response[1] != 0x00 {
		return errors.NewError(errors.ErrProxy, "认证失败", nil)
	}

	c.logger.Info("[SOCKS5] 认证成功")
	return nil
}

func (c *SOCKS5ProxyConnector) connectTarget(conn net.Conn, targetAddr string) error {
	c.logger.Info(fmt.Sprintf("[SOCKS5] 请求连接到 %s", targetAddr))

	host, port, err := net.SplitHostPort(targetAddr)
	if err != nil {
		return errors.NewError(errors.ErrProxy, "无效的目标地址", err)
	}

	// 解析端口
	portNum, err := net.LookupPort("tcp", port)
	if err != nil {
		return errors.NewError(errors.ErrProxy, "无效的端口号", err)
	}

	// 准备请求
	request := make([]byte, 4) // 版本(1) + 命令(1) + 保留(1) + 地址类型(1)
	request[0] = socks5Version
	request[1] = cmdConnect
	request[2] = 0x00 // 保留字节

	// 处理不同类型的地址
	ip := net.ParseIP(host)
	if ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			request[3] = addrTypeIPv4
			request = append(request, ip4...)
		} else {
			request[3] = addrTypeIPv6
			request = append(request, ip...)
		}
	} else {
		request[3] = addrTypeDomain
		request = append(request, byte(len(host)))
		request = append(request, host...)
	}

	// 添加端口
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(portNum))
	request = append(request, portBytes...)

	// 发送请求
	if _, err := conn.Write(request); err != nil {
		c.logger.Error("发送连接请求失败", err)
		return errors.NewError(errors.ErrProxy, "发送连接请求失败", err)
	}

	// 读取响应
	response := make([]byte, 4)
	if _, err := io.ReadFull(conn, response); err != nil {
		c.logger.Error("读取连接响应失败", err)
		return errors.NewError(errors.ErrProxy, "读取连接响应失败", err)
	}

	if response[1] != respSucceeded {
		return errors.NewError(errors.ErrProxy, fmt.Sprintf("连接请求失败，状态码: %d", response[1]), nil)
	}

	// 跳过响应中的地址信息
	switch response[3] {
	case addrTypeIPv4:
		if _, err := io.CopyN(io.Discard, conn, 4+2); err != nil {
			return errors.NewError(errors.ErrProxy, "读取响应地址失败", err)
		}
	case addrTypeIPv6:
		if _, err := io.CopyN(io.Discard, conn, 16+2); err != nil {
			return errors.NewError(errors.ErrProxy, "读取响应地址失败", err)
		}
	case addrTypeDomain:
		domainLen := make([]byte, 1)
		if _, err := io.ReadFull(conn, domainLen); err != nil {
			return errors.NewError(errors.ErrProxy, "读取域名长度失败", err)
		}
		if _, err := io.CopyN(io.Discard, conn, int64(domainLen[0])+2); err != nil {
			return errors.NewError(errors.ErrProxy, "读取响应域名失败", err)
		}
	}

	return nil
}
