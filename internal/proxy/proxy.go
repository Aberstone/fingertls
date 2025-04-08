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
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/elazarl/goproxy"

	"github.com/aberstone/tls_mitm_server/internal/errors"
)

// Proxy 代理服务器实现
type Proxy struct {
	opts *Options

	httpProxy    *goproxy.ProxyHttpServer
	listener     net.Listener
	ctx          context.Context
	cancel       context.CancelFunc
	httpHandler  RequestHandler
	httpsHandler RequestHandler
}

// NewProxy 创建新的代理服务器
func NewProxy(opts ...Option) (*Proxy, error) {
	// 应用默认选项
	options := DefaultOptions()
	// 应用自定义选项
	for _, opt := range opts {
		opt(options)
	}

	// 验证选项
	if err := validateOptions(options); err != nil {
		return nil, err
	}

	// 创建根上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 创建上游代理transport
	var transport *http.Transport
	if options.UpstreamProxy != nil {
		transport = &http.Transport{
			Proxy: http.ProxyURL(options.UpstreamProxy),
		}
	} else {
		transport = &http.Transport{}
	}

	proxy := &Proxy{
		opts:   options,
		ctx:    ctx,
		cancel: cancel,
	}

	// 初始化请求处理器
	proxy.httpHandler = NewHTTPHandler(transport, options.Logger)
	proxy.httpsHandler = NewHTTPSHandler(options.TLSDialer, options.Logger)

	// 初始化HTTP代理
	if err := proxy.initHTTPProxy(); err != nil {
		return nil, err
	}

	return proxy, nil
}

// validateOptions 验证选项
func validateOptions(opts *Options) error {
	if opts.Port <= 0 || opts.Port > 65535 {
		return errors.NewError(errors.ErrConfiguration, "无效的端口号", nil)
	}

	if opts.CACert == nil || opts.CAKey == nil {
		return errors.NewError(errors.ErrConfiguration, "未提供CA证书和私钥", nil)
	}

	return nil
}

// initHTTPProxy 初始化HTTP代理
func (p *Proxy) initHTTPProxy() error {
	p.httpProxy = goproxy.NewProxyHttpServer()

	// 使用自定义logger替换goproxy的默认logger
	p.httpProxy.Logger = newGoproxyLogger(p.opts.Logger)

	p.httpProxy.Verbose = p.opts.Verbose

	// 加载CA证书
	ca, err := tls.X509KeyPair(p.opts.CACert, p.opts.CAKey)
	if err != nil {
		return errors.NewError(errors.ErrCertificate, "加载CA证书失败", err)
	}

	goproxy.GoproxyCa = ca
	p.httpProxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)

	// 请求拦截器
	p.httpProxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		var resp *http.Response
		var err error

		// 根据请求类型选择处理器
		if req.URL.Scheme == "https" || req.Method == "CONNECT" {
			resp, err = p.httpsHandler.HandleRequest(req)
		} else {
			resp, err = p.httpHandler.HandleRequest(req)
		}

		if err != nil {
			p.opts.Logger.Error("请求处理失败", err)
			return req, goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusInternalServerError, "Internal Server Error")
		}

		// 处理响应的Content-Encoding
		if resp != nil && resp.Body != nil {
			encoding := resp.Header.Get("Content-Encoding")
			if encoding != "" {
				var reader io.ReadCloser
				var err error

				// 读取原始响应体
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					p.opts.Logger.Error("读取响应体失败", err)
					return req, resp
				}
				resp.Body.Close()

				// 根据不同的编码类型进行解码
				switch strings.ToLower(encoding) {
				case "gzip":
					reader, err = gzip.NewReader(bytes.NewReader(body))
				case "deflate":
					reader = flate.NewReader(bytes.NewReader(body))
				case "br":
					reader = io.NopCloser(brotli.NewReader(bytes.NewReader(body)))
				default:
					// 不支持的编码类型，保持原样返回
					resp.Body = io.NopCloser(bytes.NewReader(body))
					return req, resp
				}

				if err != nil {
					p.opts.Logger.Error("创建解码器失败", err)
					resp.Body = io.NopCloser(bytes.NewReader(body))
					return req, resp
				}

				// 读取解码后的内容
				decodedBody, err := io.ReadAll(reader)
				reader.Close()
				if err != nil {
					p.opts.Logger.Error("解码响应体失败", err)
					resp.Body = io.NopCloser(bytes.NewReader(body))
					return req, resp
				}

				// 移除Content-Encoding头,因为内容已经被解码
				resp.Header.Del("Content-Encoding")
				// 更新Content-Length
				resp.ContentLength = int64(len(decodedBody))
				resp.Header.Set("Content-Length", fmt.Sprint(len(decodedBody)))
				// 设置解码后的响应体
				resp.Body = io.NopCloser(bytes.NewReader(decodedBody))
			}
		}

		return req, resp
	})

	p.httpProxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		p.opts.Logger.Info(fmt.Sprintf("[MITM_SERVER] 响应: %d %s", resp.StatusCode, resp.Request.URL.String()))
		return resp
	})

	return nil
}

// Start 启动代理服务器
func (p *Proxy) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", p.opts.Port)

	p.opts.Logger.Info(fmt.Sprintf("[MITM_SERVER] 开始监听端口: %d", p.opts.Port))

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.NewError(errors.ErrNetwork, "启动监听失败", err)
	}
	p.listener = listener

	// 启动HTTP代理服务器
	go http.Serve(listener, p.httpProxy)

	return nil
}

// Stop 停止代理服务器
func (p *Proxy) Stop(ctx context.Context) error {
	p.opts.Logger.Info("[MITM_SERVER] 停止代理服务器")

	p.cancel()
	if p.listener != nil {
		return p.listener.Close()
	}
	return nil
}

// GetProxyURL 获取代理URL
func (p *Proxy) GetProxyURL() *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("localhost:%d", p.opts.Port),
	}
}
