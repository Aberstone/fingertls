package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"

	"tls_mitm_server/internal/errors"
	"tls_mitm_server/internal/interfaces"
)

// Generator CA证书生成器
type Generator struct {
	config *Config
	parent *x509.Certificate
	priv   *rsa.PrivateKey
	serial *big.Int
}

// Config 证书生成配置
type Config struct {
	Organization string
	Country      string
	CommonName   string
	ValidYears   int
}

var _ interfaces.CertificateGenerator = (*Generator)(nil)

// NewGenerator 创建新的证书生成器
func NewGenerator(config *Config) (*Generator, error) {
	if config == nil {
		return nil, errors.NewError(errors.ErrConfiguration, "证书生成器配置不能为空", nil)
	}

	if config.ValidYears == 0 {
		config.ValidYears = 10
	}

	// 生成序列号
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, errors.NewError(errors.ErrCertificate, "生成证书序列号失败", err)
	}

	return &Generator{
		config: config,
		serial: serial,
	}, nil
}

// GenerateCA 生成CA证书和私钥
func (g *Generator) GenerateCA(certPath, keyPath string) error {
	// 生成RSA私钥
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return errors.NewError(errors.ErrCertificate, "生成私钥失败", err)
	}

	// 准备证书模板
	notBefore := time.Now()
	notAfter := notBefore.AddDate(g.config.ValidYears, 0, 0)

	template := x509.Certificate{
		SerialNumber: g.serial,
		Subject: pkix.Name{
			Organization: []string{g.config.Organization},
			Country:      []string{g.config.Country},
			CommonName:   g.config.CommonName,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
	}

	// 创建自签名CA证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return errors.NewError(errors.ErrCertificate, "创建证书失败", err)
	}

	// 保存证书
	certOut, err := os.Create(certPath)
	if err != nil {
		return errors.NewError(errors.ErrCertificate, "无法创建证书文件", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return errors.NewError(errors.ErrCertificate, "无法写入证书", err)
	}

	// 保存私钥
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.NewError(errors.ErrCertificate, "无法创建私钥文件", err)
	}
	defer keyOut.Close()

	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}); err != nil {
		return errors.NewError(errors.ErrCertificate, "无法写入私钥", err)
	}

	// 保存生成的私钥和证书用于后续签发服务器证书
	g.priv = priv
	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return errors.NewError(errors.ErrCertificate, "解析CA证书失败", err)
	}
	g.parent = cert

	return nil
}

// GenerateCert 生成服务器证书
func (g *Generator) GenerateCert(domain string) error {
	if g.parent == nil || g.priv == nil {
		return errors.NewError(errors.ErrCertificate, "请先生成CA证书", nil)
	}

	// 生成新的序列号
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return errors.NewError(errors.ErrCertificate, "生成证书序列号失败", err)
	}

	// 生成服务器私钥
	serverPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return errors.NewError(errors.ErrCertificate, "生成服务器私钥失败", err)
	}

	// 准备服务器证书模板
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{g.config.Organization},
			CommonName:   domain,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // 服务器证书有效期1年
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
		DNSNames:              []string{domain},
	}

	// 使用CA证书签发服务器证书
	derBytes, err := x509.CreateCertificate(rand.Reader, template, g.parent, &serverPriv.PublicKey, g.priv)
	if err != nil {
		return errors.NewError(errors.ErrCertificate, "签发服务器证书失败", err)
	}

	// 将证书和私钥转换为PEM格式
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if certPEM == nil {
		return errors.NewError(errors.ErrCertificate, "编码服务器证书失败", nil)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPriv),
	})
	if privPEM == nil {
		return errors.NewError(errors.ErrCertificate, "编码服务器私钥失败", nil)
	}

	return nil
}
