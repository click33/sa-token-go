package adapter

// Codec defines serialization behavior interface Codec 定义序列化行为接口
type Codec interface {
	// Name returns codec name Name 返回序列化器名称
	Name() string
	// Encode encodes object to bytes Encode 将对象编码为字节数组
	Encode(v any) ([]byte, error)
	// Decode decodes bytes into target object Decode 将字节数组解码到目标对象
	Decode(data []byte, v any) error
}
