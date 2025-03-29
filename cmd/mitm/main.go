package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"tls_mitm_server/internal/config"
	"tls_mitm_server/internal/fingerprint"
	"tls_mitm_server/internal/logging"
	"tls_mitm_server/internal/proxy"
	"tls_mitm_server/internal/transport"

	utls "github.com/refraction-networking/utls"
)

var (
	// 代理配置
	port     = flag.Int("port", 8080, "代理服务器监听端口")
	upstream = flag.String("upstream", "", "上游代理URL(例如: http://proxy.example.com:8080)")

	// TLS配置
	caCert = flag.String("ca-cert", "ca.crt", "CA证书路径")
	caKey  = flag.String("ca-key", "ca.key", "CA私钥路径")
	fp     = flag.String("fingerprint", "default", "TLS指纹类型 (http1, http2)")

	// 日志配置
	logLevel  = flag.String("log-level", "info", "日志级别 (debug, info, warn, error)")
	logFormat = flag.String("log-format", "text", "日志格式 (text, json)")
	verbose   = flag.Bool("verbose", false, "显示详细日志")
)

func main() {
	flag.Parse()

	// 加载配置
	cfg := &config.Config{
		Proxy: config.ProxyConfig{
			Port:          *port,
			UpstreamProxy: *upstream,
		},
		Log: config.LogConfig{
			Level:   *logLevel,
			Format:  *logFormat,
			Output:  "stdout",
			Verbose: *verbose,
		},
	}

	// 初始化日志
	logger, err := logging.NewLogger(&cfg.Log)
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// 读取证书文件
	caCertPEM, err := os.ReadFile(*caCert)
	if err != nil {
		logger.Error("读取CA证书失败", err)
		os.Exit(1)
	}

	caKeyPEM, err := os.ReadFile(*caKey)
	if err != nil {
		logger.Error("读取CA私钥失败", err)
		os.Exit(1)
	}

	// 解析上游代理URL
	var upstreamURL *url.URL
	if *upstream != "" {
		upstreamURL, err = url.Parse(*upstream)
		if err != nil {
			logger.Error("解析上游代理URL失败", err)
			os.Exit(1)
		}
	}

	// 根据指纹类型选择对应的ClientHello规范
	var getClientHelloSpec func() *utls.ClientHelloSpec
	switch strings.ToLower(*fp) {
	case "http1":
		getClientHelloSpec = fingerprint.GetOnlyHTTP1ClientHelloSpec
		logger.Info("使用仅支持HTTP/1.1的TLS指纹")
	case "http2":
		getClientHelloSpec = fingerprint.GetOnlyHTTP2ClientHelloSpec
		logger.Info("使用仅支持HTTP/2的TLS指纹")
	default:
		getClientHelloSpec = fingerprint.GetDefaultClientHelloSpec
		logger.Info("使用默认TLS指纹")
	}

	// 创建TLS拨号器配置
	dialerConfig := transport.DialerConfig{
		GetHelloSpec:  getClientHelloSpec,
		Timeout:       30 * time.Second,
		UpstreamProxy: upstreamURL,
	}

	// 创建TLS拨号器
	dialer := transport.NewTLSDialer(dialerConfig, logger)

	// 创建代理服务器选项
	proxyOpts := []proxy.Option{
		proxy.WithPort(*port),
		proxy.WithLogger(logger),
		proxy.WithCACert(caCertPEM, caKeyPEM),
		proxy.WithTLSDialer(dialer),
		proxy.WithVerbose(*verbose),
	}

	// 创建代理服务器
	proxyServer, err := proxy.NewProxy(proxyOpts...)
	if err != nil {
		logger.Error("创建代理服务器失败", err)
		os.Exit(1)
	}

	// 启动代理服务器
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := proxyServer.Start(ctx); err != nil {
		logger.Error("启动代理服务器失败", err)
		os.Exit(1)
	}

	logger.Info("代理服务器已启动")
	logger.Info(fmt.Sprintf("监听地址: %s", proxyServer.GetProxyURL().String()))
	logger.Info(fmt.Sprintf("使用TLS指纹: %s", *fp))

	// 等待信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("正在关闭代理服务器...")
	if err := proxyServer.Stop(ctx); err != nil {
		logger.Error("关闭代理服务器失败", err)
	}
}
