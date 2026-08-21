package pool

import "sync"

var pixPool = sync.Pool{
        New: func() any {
                b := make([]byte, 0, 4096)
                return &b
        },
}

// Get returns a byte slice with length n (reusing pooled capacity when possible).
func Get(n int) []byte {
        if n <= 0 {
                return nil
        }
        bp := pixPool.Get().(*[]byte)
        buf := *bp
        if cap(buf) < n {
                buf = make([]byte, n)
        } else {
                buf = buf[:n]
                for i := range buf {
                        buf[i] = 0
                }
        }
        return buf
}

// Put returns a buffer to the pool. The slice must not be used after Put.
func Put(b []byte) {
        if b == nil || cap(b) == 0 {
                return
        }
        b = b[:0]
        pixPool.Put(&b)
}

// CapHint reports current pool new-allocation size hint (for tests/metrics).
func CapHint() int { return 4096 }
