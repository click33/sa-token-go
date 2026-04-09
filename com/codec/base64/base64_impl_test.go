// @Author daixk 2026/4/8 10:00:00
package base64

import (
	"encoding/base64"
	"reflect"
	"testing"
)

// TestBase64Serializer_Name tests serializer name behavior 测试序列化器名称行为
func TestBase64Serializer_Name(t *testing.T) {
	s := NewBase64Serializer()
	if got := s.Name(); got != "base64" {
		t.Errorf("Name() = %q, want %q", got, "base64")
	}
}

// TestBase64Serializer_Encode tests Base64 encoding behavior 测试 Base64 编码行为
func TestBase64Serializer_Encode(t *testing.T) {
	s := NewBase64Serializer()

	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{"basic struct", struct{ Name string }{"Alice"}, "eyJOYW1lIjoiQWxpY2UifQ==", false},
		{"map", map[string]int{"age": 30}, "eyJhZ2UiOjMwfQ==", false},
		{"slice", []int{1, 2, 3}, "WzEsMiwzXQ==", false},
		{"invalid type (chan)", make(chan int), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.Encode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Encode() serror = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && string(got) != tt.want {
				t.Errorf("Encode() = %q, want %q", string(got), tt.want)
			}
		})
	}
}

// TestBase64Serializer_Decode tests Base64 decoding behavior 测试 Base64 解码行为
func TestBase64Serializer_Decode(t *testing.T) {
	s := NewBase64Serializer()

	// Person defines test data for Base64 serializer tests 定义 Base64 序列化测试数据
	type Person struct {
		Name string
	}

	tests := []struct {
		name      string
		data      []byte
		targetPtr any
		want      any
		wantErr   bool
	}{
		{
			name:      "decode to struct",
			data:      []byte("eyJOYW1lIjoiQm9iIn0="),
			targetPtr: &Person{},
			want:      &Person{Name: "Bob"},
		},
		{
			name:      "decode to map",
			data:      []byte("eyJzY29yZSI6OTV9"),
			targetPtr: &map[string]int{},
			want:      &map[string]int{"score": 95},
		},
		{
			name:      "decode to slice",
			data:      []byte("WzQsNSw2XQ=="),
			targetPtr: &[]int{},
			want:      &[]int{4, 5, 6},
		},
		{
			name:      "malformed base64",
			data:      []byte("%%%invalid-base64%%%"),
			targetPtr: &struct{}{},
			wantErr:   true,
		},
		{
			name:      "malformed json after base64 decode",
			data:      []byte(base64.StdEncoding.EncodeToString([]byte("{invalid}"))),
			targetPtr: &struct{}{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.Decode(tt.data, tt.targetPtr)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Decode() serror = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.targetPtr, tt.want) {
				t.Errorf("Decode() got = %v, want %v", tt.targetPtr, tt.want)
			}
		})
	}
}
