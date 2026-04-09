// @Author daixk 2026/1/28 14:44:00
package base64

import (
	"encoding/base64"
	"encoding/json"
)

// Base64Serializer implements a Base64 JSON serializer Base64 JSON 序列化器实现
type Base64Serializer struct{}

// Encode serializes JSON and then encodes it with Base64 先编码为 JSON 再做 Base64 编码
func (s *Base64Serializer) Encode(v any) ([]byte, error) {
	// Serialize the value to JSON first 先用 JSON 序列化
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	// Encode the JSON bytes with Base64 再用 Base64 编码（得到字符串）
	b64Str := base64.StdEncoding.EncodeToString(jsonBytes)
	// Return the Base64 string as bytes 返回 []byte(b64Str)
	return []byte(b64Str), nil
}

// Decode decodes Base64 data and then deserializes JSON 先做 Base64 解码再从 JSON 反序列化
func (s *Base64Serializer) Decode(data []byte, v any) error {
	// Treat data as a Base64 string and decode it first 将 data 视为 Base64 字符串，先解码
	jsonBytes, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return err
	}
	// Deserialize the decoded bytes from JSON 再用 JSON 反序列化
	return json.Unmarshal(jsonBytes, v)
}

// Name returns the serializer name 返回序列化器名称
func (s *Base64Serializer) Name() string {
	return "base64"
}

// NewBase64Serializer creates a Base64 serializer 创建 Base64 序列化器
func NewBase64Serializer() *Base64Serializer {
	return &Base64Serializer{}
}
