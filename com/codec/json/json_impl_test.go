// @Author daixk 2026/1/21 13:48:00
package json

import (
	"reflect"
	"testing"
)

// TestJSONSerializer_Name tests serializer name behavior 测试序列化器名称行为
func TestJSONSerializer_Name(t *testing.T) {
	s := NewJSONSerializer()
	if got := s.Name(); got != "json" {
		t.Errorf("Name() = %q, want %q", got, "json")
	}
}

// TestJSONSerializer_Encode tests JSON encoding behavior 测试 JSON 编码行为
func TestJSONSerializer_Encode(t *testing.T) {
	s := NewJSONSerializer()

	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{"basic struct", struct{ Name string }{"Alice"}, `{"Name":"Alice"}`, false},
		{"map", map[string]int{"age": 30}, `{"age":30}`, false},
		{"slice", []int{1, 2, 3}, `[1,2,3]`, false},
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

// TestJSONSerializer_Decode tests JSON decoding behavior 测试 JSON 解码行为
func TestJSONSerializer_Decode(t *testing.T) {
	s := NewJSONSerializer()

	// Person defines test data for JSON serializer tests 定义 JSON 序列化测试数据
	type Person struct {
		Name string
	}

	tests := []struct {
		name      string
		data      string
		targetPtr any
		want      any
		wantErr   bool
	}{
		{
			name:      "decode to struct",
			data:      `{"Name":"Bob"}`,
			targetPtr: &Person{},
			want:      &Person{Name: "Bob"},
		},
		{
			name:      "decode to map",
			data:      `{"score":95}`,
			targetPtr: &map[string]int{},
			want:      &map[string]int{"score": 95},
		},
		{
			name:      "decode to slice",
			data:      `[4,5,6]`,
			targetPtr: &[]int{},
			want:      &[]int{4, 5, 6},
		},
		{
			name:      "malformed JSON",
			data:      `{invalid}`,
			targetPtr: &struct{}{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.Decode([]byte(tt.data), tt.targetPtr)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Decode() serror = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.targetPtr, tt.want) {
				t.Errorf("Decode() got = %v, want %v", tt.targetPtr, tt.want)
			}
		})
	}
}
