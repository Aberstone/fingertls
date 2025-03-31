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
package logging

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"

	"tls_mitm_server/internal/config"
)

// Logger 封装日志记录器
type Logger struct {
	logger zerolog.Logger
}

// NewLogger 创建新的日志记录器
func NewLogger(cfg *config.LogConfig) (*Logger, error) {
	// 配置输出
	var output io.Writer
	switch cfg.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// 确保日志目录存在
		if err := os.MkdirAll(filepath.Dir(cfg.Output), 0755); err != nil {
			return nil, fmt.Errorf("创建日志目录失败: %v", err)
		}
		file, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("打开日志文件失败: %v", err)
		}
		output = file
	}

	// 配置日志级别
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// 配置输出格式
	var writer io.Writer
	if cfg.Format == "json" {
		writer = output
	} else {
		writer = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
		}
	}

	// 创建日志记录器
	logger := zerolog.New(writer).
		Level(level).
		With().
		Timestamp().
		Logger()

	return &Logger{logger: logger}, nil
}

// WithContext 创建带有上下文值的新记录器
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{
		logger: l.logger.With().Logger(),
	}
}

// WithField 添加字段
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With().Interface(key, value).Logger(),
	}
}

// Debug 记录调试级别日志
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Info 记录信息级别日志
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Warn 记录警告级别日志
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Error 记录错误级别日志
func (l *Logger) Error(msg string, err error) {
	l.logger.Error().Err(err).Msg(msg)
}

// WithError 记录带有错误的日志
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		logger: l.logger.With().Err(err).Logger(),
	}
}

// WithFields 添加多个字段
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	logCtx := l.logger.With()
	for k, v := range fields {
		logCtx = logCtx.Interface(k, v)
	}
	return &Logger{
		logger: logCtx.Logger(),
	}
}

// 访问日志结构
type AccessLog struct {
	Protocol   string        // 协议
	Method     string        // HTTP方法
	URL        string        // 请求URL
	StatusCode int           // 状态码
	Duration   time.Duration // 处理时间
	BytesSent  int64         // 发送字节数
	BytesRecv  int64         // 接收字节数
	RemoteAddr string        // 远程地址
	UserAgent  string        // User-Agent
	Error      error         // 错误信息
}

// LogAccess 记录访问日志
func (l *Logger) LogAccess(log AccessLog) {
	event := l.logger.Info().
		Str("type", "access").
		Str("protocol", log.Protocol).
		Str("method", log.Method).
		Str("url", log.URL).
		Int("status", log.StatusCode).
		Dur("duration", log.Duration).
		Int64("bytes_sent", log.BytesSent).
		Int64("bytes_recv", log.BytesRecv).
		Str("remote_addr", log.RemoteAddr)

	if log.UserAgent != "" {
		event = event.Str("user_agent", log.UserAgent)
	}

	if log.Error != nil {
		event = event.Err(log.Error)
	}

	event.Send()
}

// Global 全局日志实例
var Global *Logger

// InitGlobal 初始化全局日志实例
func InitGlobal(cfg *config.LogConfig) error {
	logger, err := NewLogger(cfg)
	if err != nil {
		return err
	}
	Global = logger
	return nil
}
