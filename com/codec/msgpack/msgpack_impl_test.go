// @Author daixk 2026/1/21 13:48:00
package msgpack

import (
	"reflect"
	"testing"
)

// TestMsgPackSerializer_Name tests serializer name behavior 测试序列化器名称行为
func TestMsgPackSerializer_Name(t *testing.T) {
	s := NewMsgPackSerializer()
	if got := s.Name(); got != "msgpack" {
		t.Errorf("Name() = %q, want %q", got, "msgpack")
	}
}

// TestMsgPackSerializer_Encode tests MsgPack encoding behavior 测试 MsgPack 编码行为
func TestMsgPackSerializer_Encode(t *testing.T) {
	s := NewMsgPackSerializer()

	// Person defines test data for MsgPack serializer tests 定义 MsgPack 序列化测试数据
	type Person struct {
		Name string
		Age  int
	}

	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name:  "basic struct",
			input: Person{Name: "Alice", Age: 30},
		},
		{
			name:  "map",
			input: map[string]int{"score": 95},
		},
		{
			name:  "slice",
			input: []int{1, 2, 3},
		},
		{
			name:  "primitive",
			input: 42,
		},
		{
			name:    "unsupported type (chan)",
			input:   make(chan int),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.Encode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() serror = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestMsgPackSerializer_Decode tests MsgPack decoding behavior 测试 MsgPack 解码行为
func TestMsgPackSerializer_Decode(t *testing.T) {
	s := NewMsgPackSerializer()

	// Person defines test data for MsgPack serializer tests 定义 MsgPack 序列化测试数据
	type Person struct {
		Name string
		Age  int
	}

	tests := []struct {
		name    string
		prepare func() ([]byte, any) // 返回编码后的数据 + 目标指针
		want    any
		wantErr bool
	}{
		{
			name: "decode to struct",
			prepare: func() ([]byte, any) {
				data, _ := s.Encode(Person{Name: "Bob", Age: 25})
				return data, &Person{}
			},
			want: &Person{Name: "Bob", Age: 25},
		},
		{
			name: "decode to map",
			prepare: func() ([]byte, any) {
				data, _ := s.Encode(map[string]int{"count": 100})
				target := &map[string]int{}
				return data, target
			},
			want: &map[string]int{"count": 100},
		},
		{
			name: "decode to slice",
			prepare: func() ([]byte, any) {
				data, _ := s.Encode([]string{"a", "b"})
				target := &[]string{}
				return data, target
			},
			want: &[]string{"a", "b"},
		},
		{
			name: "malformed data",
			prepare: func() ([]byte, any) {
				return []byte{0xFF, 0xFF, 0xFF}, &struct{}{} // 无效 msgpack 数据
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, targetPtr := tt.prepare()
			err := s.Decode(data, targetPtr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() serror = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(targetPtr, tt.want) {
				t.Errorf("Decode() got = %v, want %v", targetPtr, tt.want)
			}
		})
	}
}
