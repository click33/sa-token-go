// @Author daixk 2025/11/27 22:58:00
package gf

import (
	"context"
	"github.com/gogf/gf/v2/os/glog"
)

// GFLogger implements the GoFrame logger adapter GoFrame 日志适配器实现
type GFLogger struct {
	ctx context.Context
	l   *glog.Logger
}

// NewGFLogger creates a GoFrame logger adapter 创建新的 GoFrame 日志适配器
func NewGFLogger(ctx context.Context, l *glog.Logger) *GFLogger {
	return &GFLogger{
		ctx: ctx,
		l:   l,
	}
}

// Print writes plain logs through GoFrame 通过 GoFrame 输出普通日志
func (g *GFLogger) Print(v ...any) {
	g.l.Print(g.ctx, v...)
}

// Printf writes formatted logs through GoFrame 通过 GoFrame 输出格式化日志
func (g *GFLogger) Printf(format string, v ...any) {
	g.l.Printf(g.ctx, format, v...)
}

// Debug writes debug logs through GoFrame 通过 GoFrame 输出调试日志
func (g *GFLogger) Debug(v ...any) {
	g.l.Debug(g.ctx, v...)
}

// Debugf writes formatted debug logs through GoFrame 通过 GoFrame 输出格式化调试日志
func (g *GFLogger) Debugf(format string, v ...any) {
	g.l.Debugf(g.ctx, format, v...)
}

// Info writes info logs through GoFrame 通过 GoFrame 输出信息日志
func (g *GFLogger) Info(v ...any) {
	g.l.Info(g.ctx, v...)
}

// Infof writes formatted info logs through GoFrame 通过 GoFrame 输出格式化信息日志
func (g *GFLogger) Infof(format string, v ...any) {
	g.l.Infof(g.ctx, format, v...)
}

// Warn writes warning logs through GoFrame 通过 GoFrame 输出警告日志
func (g *GFLogger) Warn(v ...any) {
	g.l.Warning(g.ctx, v...)
}

// Warnf writes formatted warning logs through GoFrame 通过 GoFrame 输出格式化警告日志
func (g *GFLogger) Warnf(format string, v ...any) {
	g.l.Warningf(g.ctx, format, v...)
}

// Error writes error logs through GoFrame 通过 GoFrame 输出错误日志
func (g *GFLogger) Error(v ...any) {
	g.l.Error(g.ctx, v...)
}

// Errorf writes formatted error logs through GoFrame 通过 GoFrame 输出格式化错误日志
func (g *GFLogger) Errorf(format string, v ...any) {
	g.l.Errorf(g.ctx, format, v...)
}
