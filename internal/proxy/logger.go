package proxy

import (
	"fmt"
	"strings"

	"tls_mitm_server/internal/logging"
)

// Logger 封装代理服务器的日志功能
type Logger interface {
	Printf(format string, v ...interface{})
}

// goproxyLoggerAdapter 实现goproxy.Logger接口的适配器
type goproxyLoggerAdapter struct {
	logger *logging.Logger
}

// newGoproxyLogger 创建新的goproxy日志适配器
func newGoproxyLogger(logger *logging.Logger) Logger {
	return &goproxyLoggerAdapter{
		logger: logger,
	}
}

// Printf 实现goproxy.Logger接口
// 将goproxy的日志转换为我们的日志格式
func (l *goproxyLoggerAdapter) Printf(format string, v ...interface{}) {
	// 移除末尾的换行符
	msg := fmt.Sprintf(format, v...)
	msg = strings.TrimRight(msg, "\n")
	msg = "[goproxy] " + msg
	l.logger.Info(msg)
}
