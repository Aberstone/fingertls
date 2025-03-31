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
package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"tls_mitm_server/internal/interfaces"
	"tls_mitm_server/internal/logging"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

// RequestHandler 请求处理器接口
type RequestHandler interface {
	// HandleRequest 处理HTTP/HTTPS请求
	HandleRequest(req *http.Request) (*http.Response, error)
}

// HTTPHandler 处理HTTP请求
type HTTPHandler struct {
	transport *http.Transport
	logger    *logging.Logger
}

// NewHTTPHandler 创建HTTP请求处理器
func NewHTTPHandler(upstreamProxy *http.Transport, logger *logging.Logger) RequestHandler {
	return &HTTPHandler{
		transport: upstreamProxy,
		logger:    logger,
	}
}

func (h *HTTPHandler) HandleRequest(req *http.Request) (*http.Response, error) {
	h.logger.Info(fmt.Sprintf("[HTTP_PROXY] HTTP请求: %s %s", req.Method, req.URL))

	// 创建新的请求以避免修改原始请求
	newReq := &http.Request{
		Method: req.Method,
		URL:    req.URL,
		Header: req.Header.Clone(),
		Body:   req.Body,
		Host:   req.Host,
	}

	// 设置请求上下文
	newReq = newReq.WithContext(req.Context())

	// 发送请求
	client := &http.Client{Transport: h.transport}
	resp, err := client.Do(newReq)
	if err != nil {
		h.logger.Error("HTTP请求失败", err)
		return nil, err
	}

	h.logger.Info(fmt.Sprintf("[HTTP_PROXY] HTTP响应: %d %s", resp.StatusCode, req.URL))

	return resp, nil
}

// HTTPSHandler 处理HTTPS请求
type HTTPSHandler struct {
	tlsDialer interfaces.TLSDialer
	logger    *logging.Logger
}

// NewHTTPSHandler 创建HTTPS请求处理器
func NewHTTPSHandler(tlsDialer interfaces.TLSDialer, logger *logging.Logger) RequestHandler {
	return &HTTPSHandler{
		tlsDialer: tlsDialer,
		logger:    logger,
	}
}

func (h *HTTPSHandler) HandleRequest(req *http.Request) (*http.Response, error) {
	h.logger.Info(fmt.Sprintf("[HTTP_PROXY] HTTPS请求: %s %s", req.Method, req.URL))

	// 建立TLS连接
	uconn, err := h.tlsDialer.DialTLS(context.TODO(), "tcp", req.URL.Host)
	if err != nil {
		h.logger.Error("TLS连接失败", err)
		return nil, err
	}

	// 创建新的请求
	newReq := &http.Request{
		Method: req.Method,
		URL:    req.URL,
		Header: req.Header.Clone(),
		Body:   req.Body,
		Host:   req.Host,
	}

	// 创建自定义transport并发送请求
	var customTransport http.RoundTripper
	uconnTyped, ok := uconn.(*utls.UConn)
	if !ok {
		h.logger.Error("TLS连接失败: 返回的连接不是 *utls.UConn 类型", nil)
		return nil, fmt.Errorf("unexpected connection type: %T", uconn)
	}
	if uconnTyped.ConnectionState().NegotiatedProtocol == "h2" {
		newReq.Proto = "HTTP/2.0"
		newReq.ProtoMajor = 2
		newReq.ProtoMinor = 0
		customTransport = &http2.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
				return uconn, nil
			},
		}
	} else {
		newReq.Proto = "HTTP/1.1"
		newReq.ProtoMajor = 1
		newReq.ProtoMinor = 0
		customTransport = &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return uconn, nil
			},
		}
	}

	client := &http.Client{Transport: customTransport}
	resp, err := client.Do(newReq)
	if err != nil {
		h.logger.Error("HTTPS请求失败", err)
		return nil, err
	}

	h.logger.Info(fmt.Sprintf("[HTTP_PROXY] HTTPS响应: %d %s", resp.StatusCode, req.URL))

	return resp, nil
}
