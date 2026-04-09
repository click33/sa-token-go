package adapter

// Log defines logging behavior interface Log 定义日志行为接口
type Log interface {
	// Print logs plain message Print 输出无级别日志
	Print(v ...any)
	// Printf logs formatted plain message Printf 输出无级别格式化日志
	Printf(format string, v ...any)

	// Debug logs debug message Debug 输出调试级别日志
	Debug(v ...any)
	// Debugf logs formatted debug message Debugf 输出调试级别格式化日志
	Debugf(format string, v ...any)
	// Info logs info message Info 输出信息级别日志
	Info(v ...any)
	// Infof logs formatted info message Infof 输出信息级别格式化日志
	Infof(format string, v ...any)
	// Warn logs warn message Warn 输出警告级别日志
	Warn(v ...any)
	// Warnf logs formatted warn message Warnf 输出警告级别格式化日志
	Warnf(format string, v ...any)
	// Error logs error message Error 输出错误级别日志
	Error(v ...any)
	// Errorf logs formatted error message Errorf 输出错误级别格式化日志
	Errorf(format string, v ...any)
}

// LogLevel defines log level LogLevel 定义日志级别
type LogLevel int

const (
	// LogLevelDebug represents debug level LogLevelDebug 表示调试级别
	LogLevelDebug LogLevel = iota + 1
	// LogLevelInfo represents info level LogLevelInfo 表示信息级别
	LogLevelInfo
	// LogLevelWarn represents warn level LogLevelWarn 表示警告级别
	LogLevelWarn
	// LogLevelError represents error level LogLevelError 表示错误级别
	LogLevelError
)

// String returns string form of log level String 返回日志级别字符串表示
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// LogControl defines runtime log control interface LogControl 定义日志运行时控制接口
type LogControl interface {
	Log

	// Close closes logger and releases resources Close 关闭日志并释放资源
	Close()
	// Flush flushes buffered logs Flush 刷新缓冲区
	Flush()

	// SetLevel updates log level SetLevel 动态更新日志级别
	SetLevel(level LogLevel)
	// SetPrefix updates log prefix SetPrefix 动态更新日志前缀
	SetPrefix(prefix string)
	// SetStdout toggles stdout output SetStdout 开关控制台输出
	SetStdout(enable bool)

	// LogPath gets log path LogPath 获取日志目录
	LogPath() string
	// DropCount gets dropped log count DropCount 获取丢弃的日志数量
	DropCount() uint64
}
