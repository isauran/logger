package buffer

import (
	"sync"
	"sync/atomic"
)

// Pool manages a pool of byte buffers with size limits
type Pool struct {
	pool         sync.Pool
	maxSize      int64
	maxBuffers   int64
	totalBuffers atomic.Int64
}

// DefaultMaxBufferSize is the maximum size of a buffer that will be returned to the pool
const DefaultMaxBufferSize = 1 << 16 // 64KB

// DefaultMaxBuffers is the maximum number of buffers that can be stored in the pool
const DefaultMaxBuffers = 1000

// NewPool creates a new buffer pool with size limits
func NewPool(maxSize int64, maxBuffers int64) *Pool {
	if maxSize <= 0 {
		maxSize = DefaultMaxBufferSize
	}
	if maxBuffers <= 0 {
		maxBuffers = DefaultMaxBuffers
	}

	return &Pool{
		pool: sync.Pool{
			New: func() interface{} {
				b := make([]byte, 0, 4096)
				return &b
			},
		},
		maxSize:    maxSize,
		maxBuffers: maxBuffers,
	}
}

// New creates a new buffer pool with default settings
func New(maxSize int64) *Pool {
	return NewPool(maxSize, DefaultMaxBuffers)
}

// Get retrieves a buffer from the pool
func (p *Pool) Get() *[]byte {
	buf := p.pool.Get().(*[]byte)
	p.totalBuffers.Add(1)
	*buf = (*buf)[:0] // Reset length but keep capacity
	return buf
}

// Put returns a buffer to the pool if size limits allow
func (p *Pool) Put(buf *[]byte) {
	if buf == nil {
		return
	}

	// Don't store buffers that are too large
	if cap(*buf) > int(p.maxSize) {
		p.totalBuffers.Add(-1)
		return
	}

	// Don't store more buffers than the limit
	if p.totalBuffers.Load() > p.maxBuffers {
		p.totalBuffers.Add(-1)
		return
	}

	*buf = (*buf)[:0]
	p.pool.Put(buf)
}
