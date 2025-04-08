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
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func main() {
	// http 跳过 tls 验证
	// client := &http.Client{Transport: &http.Transport{
	// 	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	// },
	// }
	// respWithoutMitmTLS, _ := client.Get("https://tls.browserscan.net/api/tls")
	// defer respWithoutMitmTLS.Body.Close()
	// contentWithoutMitmTLS, _ := io.ReadAll(respWithoutMitmTLS.Body)
	// fmt.Println(string(contentWithoutMitmTLS))
	// fmt.Println("========================================")

	clientWithMitmTLS := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, Proxy: func(r *http.Request) (*url.URL, error) {
			return &url.URL{
				Scheme: "http",
				Host:   "localhost:8080",
			}, nil
		}}}
	respWithMitmTLS, _ := clientWithMitmTLS.Get("https://tls.browserscan.net/api/tls")
	defer respWithMitmTLS.Body.Close()
	contentWithMitmTLS, _ := io.ReadAll(respWithMitmTLS.Body)
	fmt.Println(string(contentWithMitmTLS))
	fmt.Println("========================================")
	return
}
