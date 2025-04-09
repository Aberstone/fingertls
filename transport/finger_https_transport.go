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
	ctls "crypto/tls"
	"net"
	"net/http"

	"github.com/aberstone/fingertls/transport/tls"
	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

type FingerHttpsTransport struct {
	dialer tls.ITLSDialer
}

func NewFingerHttpsTransport(dialer tls.ITLSDialer) *FingerHttpsTransport {
	return &FingerHttpsTransport{
		dialer: dialer,
	}
}

func (t *FingerHttpsTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	hostWithPort := req.URL.Hostname() + ":" + req.URL.Port()
	tlsConn, _ := t.dialer.DialTLS(req.Context(), "tcp", hostWithPort)
	var tripper http.RoundTripper
	switch tlsConn.(*utls.UConn).ConnectionState().NegotiatedProtocol {
	case "h2":
		req.Proto = "HTTP/2.0"
		req.ProtoMajor = 2
		req.ProtoMinor = 0
		tripper = &http2.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string, cfg *ctls.Config) (net.Conn, error) {
				return tlsConn, nil
			},
		}
	default:
		req.Proto = "HTTP/1.1"
		req.ProtoMajor = 1
		req.ProtoMinor = 0
		tripper = &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return tlsConn, nil
			},
		}
	}
	return tripper.RoundTrip(req)
}
