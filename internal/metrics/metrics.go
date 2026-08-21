package metrics

import "sync/atomic"

// Collector holds atomic counters for pipeline observability.
type Collector struct {
        runs       int64
        hits       int64
        miss       int64
        decodeFail int64
        encodeFail int64
        bytesIn    int64
        bytesOut   int64
        closed     int32
}

// Snapshot is a point-in-time view.
type Snapshot struct {
        Runs       int64
        CacheHits  int64
        CacheMiss  int64
        DecodeFail int64
        EncodeFail int64
        BytesIn    int64
        BytesOut   int64
        Closed     bool
}

func New() *Collector { return &Collector{} }

func (c *Collector) RunOK(in, out int64) {
        atomic.AddInt64(&c.runs, 1)
        atomic.AddInt64(&c.bytesIn, in)
        atomic.AddInt64(&c.bytesOut, out)
}

func (c *Collector) Hit(in int64) {
        atomic.AddInt64(&c.hits, 1)
        atomic.AddInt64(&c.bytesIn, in)
}

func (c *Collector) Miss(in int64) {
        atomic.AddInt64(&c.miss, 1)
        atomic.AddInt64(&c.bytesIn, in)
}

func (c *Collector) DecodeError() { atomic.AddInt64(&c.decodeFail, 1) }
func (c *Collector) EncodeError() { atomic.AddInt64(&c.encodeFail, 1) }

func (c *Collector) SetClosed(v bool) {
        if v {
                atomic.StoreInt32(&c.closed, 1)
        } else {
                atomic.StoreInt32(&c.closed, 0)
        }
}

func (c *Collector) Snapshot() Snapshot {
        return Snapshot{
                Runs:       atomic.LoadInt64(&c.runs),
                CacheHits:  atomic.LoadInt64(&c.hits),
                CacheMiss:  atomic.LoadInt64(&c.miss),
                DecodeFail: atomic.LoadInt64(&c.decodeFail),
                EncodeFail: atomic.LoadInt64(&c.encodeFail),
                BytesIn:    atomic.LoadInt64(&c.bytesIn),
                BytesOut:   atomic.LoadInt64(&c.bytesOut),
                Closed:     atomic.LoadInt32(&c.closed) != 0,
        }
}
