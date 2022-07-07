package logf

import (
	"strconv"
	"sync"
	"time"
)

// ref: https://github.com/VictoriaMetrics/VictoriaMetrics/blob/master/lib/bytesutil/bytebuffer.go
// byteBufferPool is a pool of byteBuffer
type byteBufferPool struct {
	p sync.Pool
}

// Get returns a new instance of byteBuffer or gets from the object pool
func (bbp *byteBufferPool) Get() *byteBuffer {
	bbv := bbp.p.Get()
	if bbv == nil {
		return &byteBuffer{}
	}
	return bbv.(*byteBuffer)
}

// Put puts back the ByteBuffer into the object pool
func (bbp *byteBufferPool) Put(bb *byteBuffer) {
	bb.Reset()
	bbp.p.Put(bb)
}

// byteBuffer is a wrapper around byte array
type byteBuffer struct {
	B []byte
}

// AppendByte appends a single byte to the buffer.
func (bb *byteBuffer) AppendByte(b byte) {
	bb.B = append(bb.B, b)
}

// AppendString appends a string to the buffer.
func (bb *byteBuffer) AppendString(s string) {
	bb.B = append(bb.B, s...)
}

// AppendInt appends an integer to the underlying buffer (assuming base 10).
func (bb *byteBuffer) AppendInt(i int64) {
	bb.B = strconv.AppendInt(bb.B, i, 10)
}

// AppendTime appends the time formatted using the specified layout.
func (bb *byteBuffer) AppendTime(t time.Time, layout string) {
	bb.B = t.AppendFormat(bb.B, layout)
}

// AppendBool appends a bool to the underlying buffer.
func (bb *byteBuffer) AppendBool(v bool) {
	bb.B = strconv.AppendBool(bb.B, v)
}

// AppendFloat appends a float to the underlying buffer.
func (bb *byteBuffer) AppendFloat(f float64, bitSize int) {
	bb.B = strconv.AppendFloat(bb.B, f, 'f', -1, bitSize)
}

// Bytes returns a mutable reference to the underlying buffer.
func (bb *byteBuffer) Bytes() []byte {
	return bb.B
}

// Reset resets the underlying buffer.
func (bb *byteBuffer) Reset() {
	bb.B = bb.B[:0]
}
