// @Author daixk 2025/11/27 21:08:00
package nop

// NopLogger implements a no-op logger 用于禁用所有日志输出的空日志器
type NopLogger struct{}

// NewNopLogger creates a no-op logger 创建一个新的空日志器
func NewNopLogger() *NopLogger {
	return &NopLogger{}
}

// Print drops plain log messages 丢弃普通日志消息
func (n *NopLogger) Print(v ...any) {}

// Printf drops formatted log messages 丢弃格式化日志消息
func (n *NopLogger) Printf(format string, v ...any) {}

// Debug drops debug log messages 丢弃调试日志消息
func (n *NopLogger) Debug(v ...any) {}

// Debugf drops formatted debug log messages 丢弃格式化调试日志消息
func (n *NopLogger) Debugf(format string, v ...any) {}

// Info drops info log messages 丢弃信息日志消息
func (n *NopLogger) Info(v ...any) {}

// Infof drops formatted info log messages 丢弃格式化信息日志消息
func (n *NopLogger) Infof(format string, v ...any) {}

// Warn drops warning log messages 丢弃警告日志消息
func (n *NopLogger) Warn(v ...any) {}

// Warnf drops formatted warning log messages 丢弃格式化警告日志消息
func (n *NopLogger) Warnf(format string, v ...any) {}

// Error drops error log messages 丢弃错误日志消息
func (n *NopLogger) Error(v ...any) {}

// Errorf drops formatted error log messages 丢弃格式化错误日志消息
func (n *NopLogger) Errorf(format string, v ...any) {}
