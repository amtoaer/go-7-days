package cache

// ByteView 对缓存中值的抽象
type ByteView struct {
	b []byte
}

func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) ByteSlice() []byte { // 返回深拷贝的[]byte
	return copySlice(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func copySlice(b []byte) []byte {
	readOnly := make([]byte, len(b))
	copy(readOnly, b)
	return readOnly
}
