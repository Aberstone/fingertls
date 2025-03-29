package main

import (
	"flag"
	"log"

	"tls_mitm_server/internal/cert"
	"tls_mitm_server/internal/errors"
)

var (
	outCert = flag.String("cert", "ca.crt", "输出CA证书文件路径")
	outKey  = flag.String("key", "ca.key", "输出CA私钥文件路径")
	org     = flag.String("org", "MITM Proxy CA", "证书组织名称")
	country = flag.String("country", "CN", "证书国家代码")
	name    = flag.String("name", "MITM Proxy CA", "证书通用名称")
	years   = flag.Int("years", 10, "证书有效期(年)")
)

func main() {
	flag.Parse()

	config := &cert.Config{
		Organization: *org,
		Country:      *country,
		CommonName:   *name,
		ValidYears:   *years,
	}

	generator, err := cert.NewGenerator(config)
	if err != nil {
		log.Fatalf("创建证书生成器失败: %v", err)
	}

	log.Println("正在生成CA证书和私钥...")
	if err := generator.GenerateCA(*outCert, *outKey); err != nil {
		if errors.IsErrorType(err, errors.ErrCertificate) {
			log.Fatalf("生成证书失败: %v", err)
		} else {
			log.Fatalf("遇到未知错误: %v", err)
		}
	}

	log.Println("CA证书和私钥已成功生成!")
	log.Println("请将CA证书安装到您的操作系统/浏览器信任的证书列表中。")
}
