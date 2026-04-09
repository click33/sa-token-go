// @Author daixk 2025/11/27 20:57:00
package json

import (
	"encoding/json"
)

// JSONSerializer implements a JSON serializer JSON 序列化器实现
type JSONSerializer struct{}

// Encode serializes a value into JSON 编码为 JSON
func (s *JSONSerializer) Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Decode deserializes JSON into a value 从 JSON 解码
func (s *JSONSerializer) Decode(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// Name returns the serializer name 返回序列化器名称
func (s *JSONSerializer) Name() string { return "json" }

// NewJSONSerializer creates a JSON serializer 创建 JSON 序列化器
func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}
