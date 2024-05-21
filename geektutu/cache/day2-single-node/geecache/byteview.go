package geecache

// A ByteView holds an immutable view of bytes
type ByteView struct {
	b []byte //b会存储真实的缓存值，选择byte类型是为了能够支持任意的数据类型的存储，例如字符串、图片等
}

// Len return the view's length
func (v ByteView) Len() int {
	return len(v.b) //要求byteview缓存对象必须实现value接口，即Len()方法，返回其所占用的内存大小
}

// ByteSlice returns a copy of the data as a byte slice
// 因为b是只读的，使用ByteSlice()方法返回一个拷贝，防止缓存值被外部程序修改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// String returns the data as a string ,make a copy if necessary
func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
