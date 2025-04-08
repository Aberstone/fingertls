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
	"context"
	ctls "crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/aberstone/tls_mitm_server/transport/tls"
	"github.com/aberstone/tls_mitm_server/transport/tls/fingerprint"
	"github.com/sergi/go-diff/diffmatchpatch"
	"golang.org/x/net/http2"
)

func main() {
	//https://tls.browserscan.net/api/tls

	dialer1 := tls.NewTLSDialer(tls.WithSpecFactory(fingerprint.GetDefaultClientHelloSpec))
	tr1 := &http2.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string, cfg *ctls.Config) (net.Conn, error) {
			return dialer1.DialTLS(ctx, network, addr)
		},
	}

	client1 := &http.Client{
		Transport: tr1,
	}
	resp1, err := client1.Get("https://tls.browserscan.net/api/tls")
	if err != nil {
		fmt.Println("Error with first request:", err)
		return
	}
	defer resp1.Body.Close()
	content1, _ := io.ReadAll(resp1.Body)

	dialer2 := tls.NewTLSDialer(tls.WithSpecFactory(fingerprint.GetOnlyHTTP2ClientHelloSpec))
	tr2 := &http2.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string, cfg *ctls.Config) (net.Conn, error) {
			return dialer2.DialTLS(ctx, network, addr)
		},
	}
	client2 := &http.Client{
		Transport: tr2}
	resp2, err := client2.Get("https://tls.browserscan.net/api/tls")
	if err != nil {
		fmt.Println("Error with second request:", err)
		return
	}
	defer resp2.Body.Close()
	content2, _ := io.ReadAll(resp2.Body)

	// Show diff using go-diff
	fmt.Println("\nDiff between responses:")
	showDiff(string(content1), string(content2))
}

// showDiff prints a colorized line-by-line diff between two strings
func showDiff(text1, text2 string) {
	// Import required at the top of the file:
	// import (
	//    "github.com/sergi/go-diff/diffmatchpatch"
	//    "strings"
	// )

	// Create a new diff instance
	dmp := diffmatchpatch.New()

	// Convert lines to a single string with special line separators
	chars1, chars2, lineArray := dmp.DiffLinesToChars(text1, text2)

	// Get the diff using the line-char representation
	diffs := dmp.DiffMain(chars1, chars2, false)

	// Convert back to text
	diffs = dmp.DiffCharsToLines(diffs, lineArray)

	// Display the diff line by line
	for _, diff := range diffs {
		diffLines := strings.Split(diff.Text, "\n")
		for i, line := range diffLines {
			// Skip empty trailing line from split
			if i == len(diffLines)-1 && line == "" {
				continue
			}

			switch diff.Type {
			case diffmatchpatch.DiffInsert:
				fmt.Printf("\033[32m+ %s\033[0m\n", line) // Green for additions
			case diffmatchpatch.DiffDelete:
				fmt.Printf("\033[31m- %s\033[0m\n", line) // Red for deletions
			case diffmatchpatch.DiffEqual:
				fmt.Printf("  %s\n", line) // Unchanged lines with two spaces prefix
			}
		}
	}
}
