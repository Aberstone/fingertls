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
package main

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	ctls "crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/aberstone/fingertls/transport/tls"
	"github.com/aberstone/fingertls/transport/tls/fingerprint"
	"github.com/andybalholm/brotli"
	"github.com/elazarl/goproxy"
	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

func newHttpTransport(upstreamProxy *url.URL) (transport *http.Transport) {
	if upstreamProxy != nil {
		transport = &http.Transport{
			Proxy: http.ProxyURL(upstreamProxy),
		}
	} else {
		transport = &http.Transport{}
	}
	return transport
}

func handleHttpRequest(rawRequest *http.Request, upstreamProxy *url.URL) (*http.Response, error) {
	newReq := &http.Request{
		Method: rawRequest.Method,
		URL:    rawRequest.URL,
		Header: rawRequest.Header.Clone(),
		Body:   rawRequest.Body,
		Host:   rawRequest.Host,
	}

	transport := newHttpTransport(upstreamProxy)

	// 设置请求上下文
	newReq = newReq.WithContext(rawRequest.Context())

	// 发送请求
	client := &http.Client{Transport: transport}
	resp, err := client.Do(newReq)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return nil, err
	}

	return resp, nil
}

func handleHttpsRequest(rawRequest *http.Request, upstreamProxy *url.URL) (*http.Response, error) {

	newReq := &http.Request{
		Method: rawRequest.Method,
		URL:    rawRequest.URL,
		Header: rawRequest.Header.Clone(),
		Body:   rawRequest.Body,
		Host:   rawRequest.Host,
	}

	dialer := tls.NewTLSDialer(
		tls.WithSpecFactory(fingerprint.GetDefaultClientHelloSpec),
		tls.WithUpstreamProxy(upstreamProxy),
		tls.WithProxyTimeout(30),
	)

	uconn, _ := dialer.DialTLS(context.TODO(), "tcp", rawRequest.URL.Hostname()+":443")

	// 创建自定义transport并发送请求
	var customTransport http.RoundTripper
	uconnTyped, ok := uconn.(*utls.UConn)
	if !ok {
		return nil, fmt.Errorf("unexpected connection type: %T", uconn)
	}
	if uconnTyped.ConnectionState().NegotiatedProtocol == "h2" {
		newReq.Proto = "HTTP/2.0"
		newReq.ProtoMajor = 2
		newReq.ProtoMinor = 0
		customTransport = &http2.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string, cfg *ctls.Config) (net.Conn, error) {
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
	fmt.Println(resp.Proto, resp.ProtoMajor, resp.ProtoMinor)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func handleContentEncoding(resp *http.Response) *http.Response {
	if resp != nil && resp.Body != nil {
		encoding := resp.Header.Get("Content-Encoding")
		if encoding != "" {
			var reader io.ReadCloser
			var err error

			// 读取原始响应体
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response body: %v\n", err)
				return resp
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
				return resp
			}

			if err != nil {
				fmt.Printf("Error creating decoder: %v\n", err)
				resp.Body = io.NopCloser(bytes.NewReader(body))
				return resp
			}

			// 读取解码后的内容
			decodedBody, err := io.ReadAll(reader)
			reader.Close()
			if err != nil {
				fmt.Printf("Error decoding response body: %v\n", err)
				resp.Body = io.NopCloser(bytes.NewReader(body))
				return resp
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
	return resp
}

func main() {
	ps := goproxy.NewProxyHttpServer()
	ps.OnRequest().HandleConnect(goproxy.AlwaysMitm)

	// 无法直接重新赋值 ps.Tr，来实现额外的功能;
	// 因为需要自定义 tls 指纹除了需要定义各项 tls 能力插件一致以外，同样需要保证提供插件的顺序，而 http.Transport 类型中，依赖的 crypto/tls 库未提供保证顺序的实现
	// 所以使用了第三方的 utls 实现，但是也带来了在 http2 下无法直接兼容 http.Transport 的问题。
	// 因为需要直接替换 http.Transport 中的 DialTLSContext 相关的实现，这部分会返回一个 net.Conn 接口，实际类型为 *utls.UConn, 而不是 *tls.Conn，所以 http.Transport 无法获取到正确的 tls.ConnectionState
	// 需要使用 http2.Transport 来实现 http2 的支持

	ps.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			// 这里可以修改请求的 Header 或者其他操作
			if req.URL.Scheme == "https" || req.Method == "CONNECT" {
				resp, err := handleHttpsRequest(req, nil)
				if err != nil {
					fmt.Printf("Error handling HTTPS request: %v\n", err)
					return req, nil
				}
				return req, handleContentEncoding(resp)
			} else {
				resp, err := handleHttpRequest(req, nil)
				if err != nil {
					fmt.Printf("Error handling HTTPS request: %v\n", err)
					return req, nil
				}
				return req, handleContentEncoding(resp)
			}
		},
	)

	conn, _ := net.Listen("tcp", ":8080")
	defer conn.Close()
	http.Serve(conn, ps)

}
