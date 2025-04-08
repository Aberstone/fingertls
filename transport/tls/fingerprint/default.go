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
package fingerprint

import (
	utls "github.com/refraction-networking/utls"
)

type SpecFactory func() *utls.ClientHelloSpec

func GetDefaultClientHelloSpec() *utls.ClientHelloSpec {
	//修改ja4、ja3
	return &utls.ClientHelloSpec{
		TLSVersMin: utls.VersionTLS12,
		TLSVersMax: utls.VersionTLS13,
		CipherSuites: []uint16{
			utls.GREASE_PLACEHOLDER,                      // GREASE前缀
			utls.TLS_AES_128_GCM_SHA256,                  // 4865
			utls.TLS_AES_256_GCM_SHA384,                  // 4866
			utls.TLS_CHACHA20_POLY1305_SHA256,            // 4867
			utls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, // 49195
			utls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,   // 49199
			utls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, // 49196
			utls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,   // 49200
			utls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,  // 52393
			utls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,    // 52392
			utls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,      // 49171
			utls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,      // 49172
			utls.TLS_RSA_WITH_AES_128_GCM_SHA256,         // 156
			utls.TLS_RSA_WITH_AES_256_GCM_SHA384,         // 157
			utls.TLS_RSA_WITH_AES_128_CBC_SHA,            // 47
			utls.TLS_RSA_WITH_AES_256_CBC_SHA,            // 53
		},
		CompressionMethods: []byte{0}, // 无压缩
		Extensions: []utls.TLSExtension{
			// 严格按照JA3扩展顺序
			&utls.SNIExtension{},                  // 0
			&utls.ExtendedMasterSecretExtension{}, // 23
			&utls.RenegotiationInfoExtension{ // 65281
				Renegotiation: utls.RenegotiateOnceAsClient},
			&utls.SupportedCurvesExtension{ // 10
				Curves: []utls.CurveID{
					utls.GREASE_PLACEHOLDER,
					utls.X25519,    // 29
					utls.CurveP256, // 23
					utls.CurveP384, // 24
				}},
			&utls.SupportedPointsExtension{ // 11
				SupportedPoints: []byte{0}}, // 0
			&utls.SessionTicketExtension{}, // 35
			&utls.ALPNExtension{ // 16
				AlpnProtocols: []string{"h2", "http/1.1"}},
			&utls.StatusRequestExtension{}, // 5
			&utls.SignatureAlgorithmsExtension{ // 13
				SupportedSignatureAlgorithms: []utls.SignatureScheme{
					utls.ECDSAWithP256AndSHA256, //0403
					utls.PSSWithSHA256,          //804
					utls.PKCS1WithSHA256,        //0401
					utls.ECDSAWithP384AndSHA384, //0503
					utls.PSSWithSHA384,          //805
					utls.PKCS1WithSHA384,        //0501
					utls.PSSWithSHA512,          //806
					utls.PKCS1WithSHA512,        //0601
				}},
			&utls.SCTExtension{}, // 18
			&utls.KeyShareExtension{ // 51
				KeyShares: []utls.KeyShare{
					{Group: utls.X25519},
					{Group: utls.CurveP256},
				}},
			&utls.PSKKeyExchangeModesExtension{ // 45
				Modes: []uint8{1}},
			&utls.SupportedVersionsExtension{ // 43
				Versions: []uint16{
					utls.GREASE_PLACEHOLDER,
					utls.VersionTLS13,
					utls.VersionTLS12}},
			&utls.UtlsCompressCertExtension{
				Algorithms: []utls.CertCompressionAlgo{
					utls.CertCompressionBrotli,
				}}, //27
			&utls.ApplicationSettingsExtension{
				SupportedProtocols: []string{"h2"},
			}, //17153
			&utls.UtlsPaddingExtension{ // 21
				GetPaddingLen: utls.BoringPaddingStyle},
			&utls.UtlsGREASEExtension{},
			&utls.UtlsGREASEExtension{},
			(utls.PreSharedKeyExtension)(&utls.FakePreSharedKeyExtension{
				Identities: []utls.PskIdentity{
					{
						Label:               []byte("identity"),
						ObfuscatedTicketAge: 0,
					},
				},
				Binders: [][]byte{make([]byte, 32)},
			}), //41
		}}
}
