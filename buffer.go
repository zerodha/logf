package logf

import (
	"strconv"
	"sync"
	"time"
)

const (
	defaultBufSize = 512
	maxBufSize     = 64 * 1024
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
		return &byteBuffer{B: make([]byte, 0, defaultBufSize)}
	}
	return bbv.(*byteBuffer)
}

// Put puts back the ByteBuffer into the object pool
func (bbp *byteBufferPool) Put(bb *byteBuffer) {
	if cap(bb.B) > maxBufSize {
		return
	}
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
	if layout == defaultTSFormat {
		bb.B = t.Truncate(time.Millisecond).AppendFormat(bb.B, time.RFC3339Nano)
		return
	}
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

// AppendUint appends an unsigned integer to the underlying buffer.
func (bb *byteBuffer) AppendUint(u uint64) {
	bb.B = strconv.AppendUint(bb.B, u, 10)
}

// AppendDuration appends a duration in the standard Go format (e.g., "1h2m3s").
func (bb *byteBuffer) AppendDuration(d time.Duration) {
	if d == 0 {
		bb.B = append(bb.B, '0', 's')
		return
	}

	if d < 0 {
		bb.B = append(bb.B, '-')
		d = -d
	}

	if d < time.Microsecond {
		bb.B = strconv.AppendInt(bb.B, int64(d), 10)
		bb.B = append(bb.B, "ns"...)
		return
	}

	if d < time.Millisecond {
		bb.appendFrac(int64(d), int64(time.Microsecond), "µs")
		return
	}

	if d < time.Second {
		bb.appendFrac(int64(d), int64(time.Millisecond), "ms")
		return
	}

	if d < time.Minute {
		bb.appendFrac(int64(d), int64(time.Second), "s")
		return
	}

	if d < time.Hour {
		mins := d / time.Minute
		bb.B = strconv.AppendInt(bb.B, int64(mins), 10)
		bb.B = append(bb.B, 'm')
		d -= mins * time.Minute
		if d > 0 {
			bb.appendFrac(int64(d), int64(time.Second), "s")
		}
		return
	}

	hours := d / time.Hour
	bb.B = strconv.AppendInt(bb.B, int64(hours), 10)
	bb.B = append(bb.B, 'h')
	d -= hours * time.Hour
	if d >= time.Minute {
		mins := d / time.Minute
		bb.B = strconv.AppendInt(bb.B, int64(mins), 10)
		bb.B = append(bb.B, 'm')
		d -= mins * time.Minute
	}
	if d > 0 {
		bb.appendFrac(int64(d), int64(time.Second), "s")
	}
}

func (bb *byteBuffer) appendFrac(v, unit int64, suffix string) {
	whole := v / unit
	frac := v % unit

	bb.B = strconv.AppendInt(bb.B, whole, 10)

	if frac > 0 {
		bb.B = append(bb.B, '.')
		scale := unit / 10
		for scale > 0 && frac > 0 {
			digit := frac / scale
			bb.B = append(bb.B, byte('0'+digit))
			frac %= scale
			scale /= 10
		}
	}
	bb.B = append(bb.B, suffix...)
}

// Bytes returns a mutable reference to the underlying buffer.
func (bb *byteBuffer) Bytes() []byte {
	return bb.B
}

// Reset resets the underlying buffer.
func (bb *byteBuffer) Reset() {
	bb.B = bb.B[:0]
}
