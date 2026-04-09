// @Author daixk 2025-12-26 14:14:15
package slog

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Logger implements ILogger and LoggerControl 日志核心实现
type Logger struct {
	cfg   *LoggerConfig // Logger configuration 日志配置
	cfgMu sync.RWMutex  // Config lock 配置锁

	fileMu     sync.Mutex // File write lock 文件写锁
	curFile    *os.File   // Current log file 当前日志文件
	curName    string     // Current file name 当前日志文件名
	curSize    int64      // Current log size 当前文件大小
	lastRotate time.Time  // Last rotation time 上次切分时间

	queue chan []byte   // Async write queue 异步写队列
	quit  chan struct{} // Stop signal 停止信号
	wg    sync.WaitGroup

	timeCache atomic.Value // Cached time info 缓存的时间信息

	closed    uint32 // Closed flag 关闭标记
	dropCount uint64 // Dropped log counter 队列满时丢弃日志计数

	closeOnce sync.Once // Ensure Close only executes once 确保 Close 只执行一次
}

// timeCacheEntry stores cached timestamps 时间缓存条目
type timeCacheEntry struct {
	sec int64  // Unix seconds Unix 秒
	str string // Formatted string 格式化字符串
}

// NewLoggerWithConfig creates a logger instance 使用配置创建日志器
func NewLoggerWithConfig(cfg *LoggerConfig) (*Logger, error) {
	newCfg, err := prepareConfig(cfg)
	if err != nil {
		return nil, err
	}

	queueSize := newCfg.QueueSize
	if queueSize <= 0 {
		queueSize = DefaultQueueSize
	}

	l := &Logger{
		cfg:        newCfg,
		queue:      make(chan []byte, queueSize),
		quit:       make(chan struct{}),
		lastRotate: time.Now(),
	}

	// Initialize time cache 初始化时间缓存
	now := time.Now()
	l.timeCache.Store(&timeCacheEntry{
		sec: now.Unix(),
		str: now.Format(newCfg.TimeFormat),
	})

	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		l.writerLoop()
	}()

	return l, nil
}

// write handles plain log output 输出普通日志
func (l *Logger) write(level LogLevel, args ...any) {
	if atomic.LoadUint32(&l.closed) != 0 {
		return
	}
	cfg := l.currentCfg()
	if level < cfg.Level {
		return
	}
	l.enqueue(l.buildLine(level, cfg, args...))
}

// writef handles formatted log output 输出格式化日志
func (l *Logger) writef(level LogLevel, format string, args ...any) {
	if atomic.LoadUint32(&l.closed) != 0 {
		return
	}
	cfg := l.currentCfg()
	if level < cfg.Level {
		return
	}
	buf := getBuf()
	_, _ = fmt.Fprintf(buf, format, args...)
	line := l.buildLine(level, cfg, buf.String())
	putBuf(buf)
	l.enqueue(line)
}

// enqueue pushes logs to the async queue 将日志推入异步队列
func (l *Logger) enqueue(b []byte) {
	if atomic.LoadUint32(&l.closed) != 0 {
		return
	}
	select {
	case l.queue <- b:
	default:
		// Drop logs when the queue is full 队列满时丢弃
		atomic.AddUint64(&l.dropCount, 1)
	}
}

// buildLine builds the complete log line 构建完整日志行
func (l *Logger) buildLine(level LogLevel, cfg LoggerConfig, args ...any) []byte {
	buf := getBuf()

	// Get cached timestamp or format a new one 获取缓存时间戳或格式化新的
	now := time.Now()
	sec := now.Unix()

	ts := l.getTimeString(now, sec, cfg.TimeFormat)
	buf.WriteString(ts)

	buf.WriteString(" [")
	buf.WriteString(levelString(level))
	buf.WriteString("] ")

	buf.WriteString(cfg.Prefix)

	for i, arg := range args {
		if i > 0 {
			buf.WriteByte(' ')
		}
		appendValue(buf, arg)
	}

	buf.WriteByte('\n')

	// Copy to a new slice to avoid buffer reuse 拷贝到新切片避免复用冲突
	out := append([]byte(nil), buf.Bytes()...)
	putBuf(buf)
	return out
}

// getTimeString returns cached or formatted time strings 返回缓存或格式化的时间字符串
func (l *Logger) getTimeString(now time.Time, sec int64, format string) string {
	// Try to load from cache 尝试从缓存加载
	if cached, ok := l.timeCache.Load().(*timeCacheEntry); ok && cached.sec == sec {
		return cached.str
	}

	// Format a new string and update cache atomically 格式化新字符串并更新缓存
	str := now.Format(format)
	l.timeCache.Store(&timeCacheEntry{sec: sec, str: str})
	return str
}

// appendValue writes a single value with optimized type handling 写入单个参数（优化类型处理）
func appendValue(buf *bytes.Buffer, v any) {
	if v == nil {
		buf.WriteString("<nil>")
		return
	}

	switch val := v.(type) {
	case string:
		buf.WriteString(val)
	case []byte:
		buf.Write(val)
	case error:
		if val != nil {
			buf.WriteString(val.Error())
		} else {
			buf.WriteString("<nil>")
		}

	// Use optimized integer handling 优化整数处理
	case int:
		buf.WriteString(strconv.FormatInt(int64(val), 10))
	case int8:
		buf.WriteString(strconv.FormatInt(int64(val), 10))
	case int16:
		buf.WriteString(strconv.FormatInt(int64(val), 10))
	case int32:
		buf.WriteString(strconv.FormatInt(int64(val), 10))
	case int64:
		buf.WriteString(strconv.FormatInt(val, 10))
	case uint:
		buf.WriteString(strconv.FormatUint(uint64(val), 10))
	case uint8:
		buf.WriteString(strconv.FormatUint(uint64(val), 10))
	case uint16:
		buf.WriteString(strconv.FormatUint(uint64(val), 10))
	case uint32:
		buf.WriteString(strconv.FormatUint(uint64(val), 10))
	case uint64:
		buf.WriteString(strconv.FormatUint(val, 10))

	case float32:
		buf.WriteString(strconv.FormatFloat(float64(val), 'g', -1, 32))
	case float64:
		buf.WriteString(strconv.FormatFloat(val, 'g', -1, 64))

	case bool:
		if val {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}

	case time.Time:
		buf.WriteString(val.Format(DefaultTimeFormat))

	default:
		_, _ = fmt.Fprint(buf, val)
	}
}

// writerLoop processes all file IO 异步写线程处理文件操作
func (l *Logger) writerLoop() {
	defer func() {
		l.Flush()
	}()

	for {
		select {
		case b, ok := <-l.queue:
			if !ok {
				return
			}
			l.writeToOutput(b)

		case <-l.quit:
			// Drain the queue before exit 退出前清空队列
			for {
				select {
				case b := <-l.queue:
					l.writeToOutput(b)
				default:
					return
				}
			}
		}
	}
}

// writeToOutput writes to file and or stdout 写入文件和/或控制台
func (l *Logger) writeToOutput(b []byte) {
	cfg := l.currentCfg()

	// Only print to console in stdout only mode 仅控制台模式
	if cfg.StdoutOnly {
		if cfg.Stdout {
			_, _ = os.Stdout.Write(b)
		}
		return
	}

	now := time.Now()

	l.fileMu.Lock()
	defer l.fileMu.Unlock()

	// Open the file when needed 无文件则打开
	if err := l.ensureLogFile(now, cfg); err != nil {
		// Fallback to stdout when file open fails 文件打开失败，回退到控制台
		if cfg.Stdout {
			_, _ = os.Stdout.Write(b)
		}
		return
	}

	if l.curFile != nil {
		n, err := l.curFile.Write(b)
		if err != nil {
			_ = l.curFile.Close()
			l.curFile = nil
			// Retry once with a new file 重试一次新文件
			if retryErr := l.openNewFile(now, cfg); retryErr == nil && l.curFile != nil {
				n, _ = l.curFile.Write(b)
				l.curSize += int64(n)
			}
		} else {
			l.curSize += int64(n)
		}
	}

	if cfg.Stdout {
		_, _ = os.Stdout.Write(b)
	}

	// Check whether rotation is needed 检测切分
	if l.shouldRotate(now, cfg) {
		_ = l.rotate(cfg)
	}
}

// ensureLogFile ensures a log file is open 确保日志文件存在
func (l *Logger) ensureLogFile(now time.Time, cfg LoggerConfig) error {
	if l.curFile == nil {
		return l.openNewFile(now, cfg)
	}
	if cfg.RotateExpire > 0 && now.Sub(l.lastRotate) >= cfg.RotateExpire {
		return l.rotate(cfg)
	}
	return nil
}

// openNewFile opens a new log file 打开新日志文件
func (l *Logger) openNewFile(now time.Time, cfg LoggerConfig) error {
	name := l.formatFileName(now, cfg)
	path := filepath.Join(cfg.Path, name)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	l.curFile = f
	l.curName = name
	l.curSize = getFileSize(f)
	l.lastRotate = now
	return nil
}

// shouldRotate checks whether rotation is needed 检查是否需要切分
func (l *Logger) shouldRotate(now time.Time, cfg LoggerConfig) bool {
	if cfg.RotateSize > 0 && l.curSize >= cfg.RotateSize {
		return true
	}
	if cfg.RotateExpire > 0 && now.Sub(l.lastRotate) >= cfg.RotateExpire {
		return true
	}
	return false
}

// rotate rotates the current log file 日志切分逻辑
func (l *Logger) rotate(cfg LoggerConfig) error {
	if l.curFile == nil {
		return nil
	}

	old := filepath.Join(cfg.Path, l.curName)
	_ = l.curFile.Sync()
	_ = l.curFile.Close()
	l.curFile = nil

	now := time.Now()
	ts := fmt.Sprintf("%s_%03d", now.Format("20060102_150405"), now.Nanosecond()/1e6)

	base := strings.TrimSuffix(l.curName, ".log")
	newName := fmt.Sprintf("%s_%s.log", base, ts)
	newPath := filepath.Join(cfg.Path, newName)

	if err := os.Rename(old, newPath); err != nil {
		// Use crypto rand for secure random numbers 使用加密安全的随机数
		randNum := secureRandomInt(1_000_000)
		_ = os.Rename(old, filepath.Join(cfg.Path, base+fmt.Sprintf("_%06d.log", randNum)))
	}

	l.curSize = 0
	l.curName = ""
	l.lastRotate = now

	// Clean up asynchronously to avoid blocking writes 异步清理避免阻塞写入
	go l.cleanup(cfg)

	return l.openNewFile(now, cfg)
}

// cleanup removes expired or extra log files 清理过期或多余日志文件
func (l *Logger) cleanup(cfg LoggerConfig) {
	// Recover from panic to avoid crashing the program 捕获 panic 避免程序崩溃
	defer func() {
		if r := recover(); r != nil {
			// Silently ignore cleanup errors 静默忽略清理错误
		}
	}()

	// base is the fixed prefix for this logger file set base 为该 Logger 对应日志文件的固定前缀
	base := normalizeBaseName(cfg.FileFormat)
	if base == "" {
		base = DefaultBaseName
	}

	files, _ := filepath.Glob(filepath.Join(cfg.Path, "*.log"))
	if len(files) == 0 {
		return
	}

	var keep []struct {
		path string
		t    time.Time
	}

	now := time.Now()
	expire := time.Time{}
	if cfg.RotateBackupDays > 0 {
		expire = now.AddDate(0, 0, -cfg.RotateBackupDays)
	}

	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}

		filename := filepath.Base(f)

		// Only handle files with the same base prefix 只处理以 base 开头的文件
		if !strings.HasPrefix(filename, base) {
			continue
		}

		// Remove expired files 清理过期文件
		if !expire.IsZero() && info.ModTime().Before(expire) {
			_ = os.Remove(f)
			continue
		}

		// Collect only backup files after rotation 当前正在写入的文件此时尚未创建（在 rotate 之后），这里收集到的全是备份文件，后续按数量进行裁剪
		keep = append(keep, struct {
			path string
			t    time.Time
		}{f, info.ModTime()})
	}

	// Limit the number of backup files by RotateBackupLimit 根据 RotateBackupLimit 限制保留的备份文件数量（不包含当前正在写的那个文件）
	if cfg.RotateBackupLimit > 0 && len(keep) > cfg.RotateBackupLimit {
		// Sort by modification time ascending 按修改时间排序，最旧的在前
		sort.Slice(keep, func(i, j int) bool { return keep[i].t.Before(keep[j].t) })

		// Delete extra backups and keep the latest cfg.RotateBackupLimit 个 删除多余的，只保留最新的 cfg.RotateBackupLimit 个
		for _, f := range keep[:len(keep)-cfg.RotateBackupLimit] {
			_ = os.Remove(f.path)
		}
	}
}

// formatFileName generates the log filename 生成日志文件名
func (l *Logger) formatFileName(t time.Time, cfg LoggerConfig) string {
	name := cfg.FileFormat
	if name == "" {
		return fmt.Sprintf("%s_%s.log", DefaultBaseName, t.Format("2006-01-02"))
	}

	r := strings.NewReplacer(
		"{Y}", t.Format("2006"),
		"{m}", t.Format("01"),
		"{d}", t.Format("02"),
	)

	name = r.Replace(name)
	if !strings.HasSuffix(name, ".log") {
		name += ".log"
	}
	return name
}

// SetLevel updates the minimum log level 动态更新日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.cfgMu.Lock()
	defer l.cfgMu.Unlock()
	if l.cfg != nil {
		l.cfg.Level = level
	}
}

// SetPrefix updates the log prefix 动态更新日志前缀
func (l *Logger) SetPrefix(prefix string) {
	l.cfgMu.Lock()
	defer l.cfgMu.Unlock()
	if l.cfg != nil {
		l.cfg.Prefix = prefix
	}
}

// SetStdout toggles stdout output 开关控制台输出
func (l *Logger) SetStdout(enable bool) {
	l.cfgMu.Lock()
	defer l.cfgMu.Unlock()
	if l.cfg != nil {
		l.cfg.Stdout = enable
	}
}

// SetConfig replaces config and reopens the log file 动态替换配置并重新创建日志文件
func (l *Logger) SetConfig(cfg *LoggerConfig) {
	newCfg, err := prepareConfig(cfg)
	if err != nil {
		return
	}

	// Lock in a consistent order fileMu then cfgMu 统一锁顺序：先 fileMu，再 cfgMu
	l.fileMu.Lock()
	defer l.fileMu.Unlock()

	l.cfgMu.Lock()
	defer l.cfgMu.Unlock()

	l.cfg = newCfg

	if l.curFile != nil {
		_ = l.curFile.Sync()
		_ = l.curFile.Close()
		l.curFile = nil
	}

	l.curName = ""
	l.curSize = 0
	l.lastRotate = time.Now()
}

// Close stops the logger 关闭日志系统
func (l *Logger) Close() {
	l.closeOnce.Do(func() {
		atomic.StoreUint32(&l.closed, 1)
		close(l.quit)

		l.wg.Wait()

		l.fileMu.Lock()
		defer l.fileMu.Unlock()

		if l.curFile != nil {
			_ = l.curFile.Sync()
			_ = l.curFile.Close()
		}
	})
}

// Flush flushes the file buffer 强制刷新文件缓冲区
func (l *Logger) Flush() {
	l.fileMu.Lock()
	defer l.fileMu.Unlock()
	if l.curFile != nil {
		_ = l.curFile.Sync()
	}
}

// LogPath returns the log directory 返回日志目录
func (l *Logger) LogPath() string {
	l.cfgMu.RLock()
	defer l.cfgMu.RUnlock()
	if l.cfg == nil {
		return ""
	}
	return l.cfg.Path
}

// DropCount returns the dropped log count 返回丢弃日志数量
func (l *Logger) DropCount() uint64 {
	return atomic.LoadUint64(&l.dropCount)
}

var bufPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

// getBuf gets a buffer from the pool 从池中获取缓冲区
func getBuf() *bytes.Buffer {
	b := bufPool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}

// putBuf returns a buffer to the pool 将缓冲区归还到池
func putBuf(b *bytes.Buffer) {
	bufPool.Put(b)
}

// getFileSize returns the file size 获取文件大小
func getFileSize(f *os.File) int64 {
	info, err := f.Stat()
	if err != nil {
		return 0
	}
	return info.Size()
}

// prepareConfig applies defaults and ensures the directory exists 应用默认配置并确保目录存在
func prepareConfig(cfg *LoggerConfig) (*LoggerConfig, error) {
	if cfg == nil {
		cfg = &LoggerConfig{}
	}

	c := *cfg // copy

	if c.TimeFormat == "" {
		c.TimeFormat = DefaultTimeFormat
	}
	if c.Prefix == "" {
		c.Prefix = DefaultPrefix
	}
	if c.QueueSize <= 0 {
		c.QueueSize = DefaultQueueSize
	}

	// Skip file config in stdout only mode 仅控制台模式不需要文件配置
	if c.StdoutOnly {
		c.Stdout = true
		return &c, nil
	}

	// Apply file related defaults in file mode 文件模式：应用文件相关默认值
	if c.FileFormat == "" {
		c.FileFormat = DefaultFileFormat
	}
	if c.RotateSize <= 0 {
		c.RotateSize = DefaultRotateSize
	}
	if c.RotateExpire < 0 {
		c.RotateExpire = 0
	}
	if c.RotateBackupLimit <= 0 {
		c.RotateBackupLimit = DefaultRotateBackupLimit
	}
	if c.RotateBackupDays < 0 {
		c.RotateBackupDays = 0
	}

	// Ensure the path exists 确保路径存在
	if c.Path == "" {
		wd, err := os.Getwd()
		if err != nil {
			wd = "."
		}
		c.Path = filepath.Join(wd, DefaultDirName)
	}

	if err := os.MkdirAll(c.Path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	return &c, nil
}

// currentCfg returns a config snapshot 返回当前配置快照
func (l *Logger) currentCfg() LoggerConfig {
	l.cfgMu.RLock()
	defer l.cfgMu.RUnlock()

	if l.cfg == nil {
		return LoggerConfig{}
	}
	return *l.cfg
}

// levelString converts log levels to strings 将日志级别转换为字符串
func levelString(level LogLevel) string {
	return level.String()
}

// normalizeBaseName extracts the static base filename 提取基础日志文件名前缀
func normalizeBaseName(format string) string {
	if format == "" {
		return DefaultBaseName
	}

	// Strip the .log suffix 去掉 .log 后缀
	name := strings.TrimSuffix(format, ".log")

	// Use the prefix before the first placeholder when placeholders exist 如果包含占位符，则取第一个占位符之前的固定前缀
	if idx := strings.Index(name, "{"); idx >= 0 {
		name = name[:idx]
		// Trim trailing separators like _ or - 去掉末尾的连接符（常见为 _ 或 -）
		name = strings.TrimRight(name, "_- ")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return DefaultBaseName
	}
	return name
}

// secureRandomInt returns a cryptographically secure random integer 返回加密安全的随机整数
func secureRandomInt(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0
	}
	return int(n.Int64())
}

// Print writes plain logs 输出普通日志
func (l *Logger) Print(v ...any) { l.write(LevelInfo, v...) }

// Printf writes formatted logs 输出格式化日志
func (l *Logger) Printf(f string, v ...any) { l.writef(LevelInfo, f, v...) }

// Debug writes debug logs 输出调试日志
func (l *Logger) Debug(v ...any) { l.write(LevelDebug, v...) }

// Debugf writes formatted debug logs 输出格式化调试日志
func (l *Logger) Debugf(f string, v ...any) { l.writef(LevelDebug, f, v...) }

// Info writes info logs 输出信息日志
func (l *Logger) Info(v ...any) { l.write(LevelInfo, v...) }

// Infof writes formatted info logs 输出格式化信息日志
func (l *Logger) Infof(f string, v ...any) { l.writef(LevelInfo, f, v...) }

// Warn writes warning logs 输出警告日志
func (l *Logger) Warn(v ...any) { l.write(LevelWarn, v...) }

// Warnf writes formatted warning logs 输出格式化警告日志
func (l *Logger) Warnf(f string, v ...any) { l.writef(LevelWarn, f, v...) }

// Error writes error logs 输出错误日志
func (l *Logger) Error(v ...any) { l.write(LevelError, v...) }

// Errorf writes formatted error logs 输出格式化错误日志
func (l *Logger) Errorf(f string, v ...any) { l.writef(LevelError, f, v...) }
