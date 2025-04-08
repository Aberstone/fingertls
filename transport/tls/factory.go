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
	"net/url"
	"time"

	"github.com/aberstone/tls_mitm_server/logging"
	"github.com/aberstone/tls_mitm_server/transport/proxy_connector"
	"github.com/aberstone/tls_mitm_server/transport/tls/fingerprint"
)

type Options struct {
	logger        logging.ILogger
	sf            fingerprint.SpecFactory
	timeout       time.Duration
	upstreamProxy *url.URL
	proxyTimeout  time.Duration
}

type Option func(*Options)

func defaultOptions() *Options {

	logger, _ := logging.NewZeroLogger(nil)

	return &Options{
		timeout:       30 * time.Second,
		upstreamProxy: nil,
		sf:            fingerprint.GetDefaultClientHelloSpec,
		logger:        logger,
		proxyTimeout:  30 * time.Second,
	}
}

func NewTLSDialer(opts ...Option) ITLSDialer {

	options := defaultOptions()

	for _, opt := range opts {
		opt(options)
	}

	var connector proxy_connector.ProxyConnector
	if options.upstreamProxy != nil {
		switch options.upstreamProxy.Scheme {
		case "http":
			connector = proxy_connector.NewHTTPProxyConnector(options.proxyTimeout, options.logger)
		case "socks5":
			connector = proxy_connector.NewSocks5ProxyConnector(options.proxyTimeout, options.logger)
		default:
			options.logger.Error("不支持的代理协议", nil)
			panic("不支持的代理协议")
		}
		return &ProxyTLSDialer{
			&BaseTLSDialer{
				opts: options,
			},
			connector,
		}
	}

	return &BaseTLSDialer{
		opts: options,
	}
}

func WithLogger(logger logging.ILogger) Option {
	return func(opts *Options) {
		opts.logger = logger
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(opts *Options) {
		opts.timeout = timeout
	}
}
func WithUpstreamProxy(upstreamProxy *url.URL) Option {
	return func(opts *Options) {
		opts.upstreamProxy = upstreamProxy
	}
}
func WithSpecFactory(sf fingerprint.SpecFactory) Option {
	return func(opts *Options) {
		opts.sf = sf
	}
}
func WithProxyTimeout(timeout time.Duration) Option {
	return func(opts *Options) {
		opts.proxyTimeout = timeout
	}
}
