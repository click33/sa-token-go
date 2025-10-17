package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
)

// RandomString generates random string of specified length | 生成指定长度的随机字符串
func RandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

// IsEmpty checks if string is empty | 检查字符串是否为空
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotEmpty checks if string is not empty | 检查字符串是否不为空
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// DefaultString returns default value if string is empty | 如果字符串为空则返回默认值
func DefaultString(s, defaultValue string) string {
	if IsEmpty(s) {
		return defaultValue
	}
	return s
}

// ContainsString checks if string slice contains item | 检查字符串数组是否包含指定字符串
func ContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveString removes item from string slice | 从字符串数组中移除指定字符串
func RemoveString(slice []string, item string) []string {
	result := make([]string, 0)
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// UniqueStrings removes duplicates from string slice | 字符串数组去重
func UniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// MergeStrings 合并多个字符串数组并去重
func MergeStrings(slices ...[]string) []string {
	result := make([]string, 0)
	for _, slice := range slices {
		result = append(result, slice...)
	}
	return UniqueStrings(result)
}

// SplitAndTrim 分割字符串并去除空格
func SplitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0)
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// GetStructTag 获取结构体字段的标签值
func GetStructTag(field reflect.StructField, tag string) string {
	return field.Tag.Get(tag)
}

// ParsePermissionTag 解析权限标签
// 格式: "perm:user:read,user:write"
func ParsePermissionTag(tag string) []string {
	if tag == "" {
		return []string{}
	}

	// 移除 "perm:" 前缀
	tag = strings.TrimPrefix(tag, "perm:")
	return SplitAndTrim(tag, ",")
}

// ParseRoleTag 解析角色标签
// 格式: "role:admin,manager"
func ParseRoleTag(tag string) []string {
	if tag == "" {
		return []string{}
	}

	// 移除 "role:" 前缀
	tag = strings.TrimPrefix(tag, "role:")
	return SplitAndTrim(tag, ",")
}

// MatchPattern 模式匹配（支持通配符*）
func MatchPattern(pattern, str string) bool {
	if pattern == "*" {
		return true
	}

	if !strings.Contains(pattern, "*") {
		return pattern == str
	}

	// 简单的通配符匹配
	parts := strings.Split(pattern, "*")
	if len(parts) == 2 {
		prefix, suffix := parts[0], parts[1]
		if prefix != "" && !strings.HasPrefix(str, prefix) {
			return false
		}
		if suffix != "" && !strings.HasSuffix(str, suffix) {
			return false
		}
		return true
	}

	return false
}

// FormatDuration 格式化时间段（秒）为人类可读格式
func FormatDuration(seconds int64) string {
	if seconds < 0 {
		return "永久"
	}

	if seconds < 60 {
		return fmt.Sprintf("%d秒", seconds)
	}

	if seconds < 3600 {
		minutes := seconds / 60
		return fmt.Sprintf("%d分钟", minutes)
	}

	if seconds < 86400 {
		hours := seconds / 3600
		return fmt.Sprintf("%d小时", hours)
	}

	days := seconds / 86400
	return fmt.Sprintf("%d天", days)
}

// ParseDuration 解析人类可读的时间段为秒
func ParseDuration(duration string) int64 {
	duration = strings.ToLower(strings.TrimSpace(duration))

	if strings.HasSuffix(duration, "s") || strings.HasSuffix(duration, "秒") {
		return parseInt64(strings.TrimSuffix(strings.TrimSuffix(duration, "s"), "秒"))
	}

	if strings.HasSuffix(duration, "m") || strings.HasSuffix(duration, "分") {
		minutes := parseInt64(strings.TrimSuffix(strings.TrimSuffix(duration, "m"), "分"))
		return minutes * 60
	}

	if strings.HasSuffix(duration, "h") || strings.HasSuffix(duration, "时") {
		hours := parseInt64(strings.TrimSuffix(strings.TrimSuffix(duration, "h"), "时"))
		return hours * 3600
	}

	if strings.HasSuffix(duration, "d") || strings.HasSuffix(duration, "天") {
		days := parseInt64(strings.TrimSuffix(strings.TrimSuffix(duration, "d"), "天"))
		return days * 86400
	}

	return parseInt64(duration)
}

func parseInt64(s string) int64 {
	var result int64
	fmt.Sscanf(s, "%d", &result)
	return result
}
