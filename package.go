// Package tls_mitm_server 实现了一个支持自定义 TLS 指纹的 MITM 代理服务器。
// 该代理服务器可以同时支持 HTTP 和 SOCKS5 协议，并且能够模拟不同浏览器的 TLS 指纹。
//
// 项目结构:
//
//	cmd/                    - 可执行程序目录
//	  ├── generate-ca/      - CA证书生成工具
//	  └── mitm/            - 代理服务器主程序
//
//	internal/              - 内部包实现
//	  ├── cert/           - 证书生成和管理
//	  ├── proxy/          - 代理服务器核心实现
//	  └── transport/      - 自定义传输层和TLS指纹实现
//
// 主要特性:
//   - 支持HTTP和SOCKS5代理协议
//   - 支持自定义TLS指纹（可模拟Chrome或Firefox）
//   - 支持上游代理链式代理
//   - 支持HTTPS解密和重加密
//   - 提供证书生成和管理工具
//
// 使用示例:
//
//	# 生成CA证书
//	$ ./gen-ca -cert ca.crt -key ca.key
//
//	# 启动HTTP代理服务器（Chrome指纹）
//	$ ./mitm -port 8080 -ca-cert ca.crt -ca-key ca.key -browser chrome -protocol http
//
//	# 启动SOCKS5代理服务器（Firefox指纹）
//	$ ./mitm -port 1080 -ca-cert ca.crt -ca-key ca.key -browser firefox -protocol socks5
//
//	# 启动双协议代理服务器（Chrome指纹）
//	$ ./mitm -port 8888 -ca-cert ca.crt -ca-key ca.key -browser chrome -protocol both
package tls_mitm_server
