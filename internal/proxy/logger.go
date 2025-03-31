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
