package imgpipe

import (
        "sync"

        "github.com/LYH2263/go-imgpipe/internal/clone"
        "github.com/LYH2263/go-imgpipe/internal/persist"
)

// Cache is the content-addressed frame cache backed by memory and optional disk.
type Cache struct {
        pipe *Pipeline
}

// Put stores a deep copy of frame under key.
func (c *Cache) Put(key string, frame *Frame) error {
        if c == nil || c.pipe == nil {
                return ErrNilPipeline
        }
        if c.pipe.Closed() {
                return ErrClosed
        }
        if key == "" || frame == nil {
                return ErrInvalidFrame
        }
        entry := &persist.Entry{
                Key:    key,
                W:      frame.W,
                H:      frame.H,
                Stride: frame.Stride,
                Format: frame.Format,
                Pixels: clone.Bytes(frame.Pixels),
                Raw:    clone.Bytes(frame.Raw),
        }
        c.pipe.mu.Lock()
        defer c.pipe.mu.Unlock()
        if c.pipe.disk != nil {
                if err := c.pipe.disk.Put(entry); err != nil {
                        return err
                }
        }
        c.pipe.mem.Put(key, entry)
        return nil
}

// Get returns a deep-copied Frame so callers cannot mutate internal entries.
func (c *Cache) Get(key string) (*Frame, bool) {
        if c == nil || c.pipe == nil || c.pipe.Closed() {
                return nil, false
        }
        c.pipe.mu.Lock()
        defer c.pipe.mu.Unlock()
        ent, ok := c.pipe.mem.Get(key)
        if !ok && c.pipe.disk != nil {
                var err error
                ent, err = c.pipe.disk.Get(key)
                if err != nil || ent == nil {
                        return nil, false
                }
                c.pipe.mem.Put(key, ent)
        }
        if ent == nil {
                return nil, false
        }
        // Always return a copy — never share Pixels with the cache entry.
        return &Frame{
                Pixels:     clone.Bytes(ent.Pixels),
                W:          ent.W,
                H:          ent.H,
                Stride:     ent.Stride,
                ColorModel: nil,
                Format:     ent.Format,
                Raw:        clone.Bytes(ent.Raw),
        }, true
}

// Has reports whether key exists in memory or disk index.
func (c *Cache) Has(key string) bool {
        if c == nil || c.pipe == nil || c.pipe.Closed() {
                return false
        }
        c.pipe.mu.Lock()
        defer c.pipe.mu.Unlock()
        if _, ok := c.pipe.mem.Get(key); ok {
                return true
        }
        if c.pipe.disk != nil {
                return c.pipe.disk.Has(key)
        }
        return false
}

// Len returns in-memory entry count.
func (c *Cache) Len() int {
        if c == nil || c.pipe == nil {
                return 0
        }
        c.pipe.mu.Lock()
        defer c.pipe.mu.Unlock()
        return c.pipe.mem.Len()
}

// Delete removes a key from memory and disk.
func (c *Cache) Delete(key string) error {
        if c == nil || c.pipe == nil {
                return ErrNilPipeline
        }
        if c.pipe.Closed() {
                return ErrClosed
        }
        c.pipe.mu.Lock()
        defer c.pipe.mu.Unlock()
        c.pipe.mem.Delete(key)
        if c.pipe.disk != nil {
                return c.pipe.disk.Delete(key)
        }
        return nil
}

// Keys returns a snapshot of memory keys.
func (c *Cache) Keys() []string {
        if c == nil || c.pipe == nil {
                return nil
        }
        c.pipe.mu.Lock()
        defer c.pipe.mu.Unlock()
        return c.pipe.mem.Keys()
}

// memIndex is a simple LRU-ish map with capacity trim.
type memIndex struct {
        mu      sync.Mutex
        limit   int
        entries map[string]*persist.Entry
        order   []string
}

func newMemIndex(limit int) *memIndex {
        return &memIndex{limit: limit, entries: make(map[string]*persist.Entry)}
}

func (m *memIndex) Put(key string, e *persist.Entry) {
        m.mu.Lock()
        defer m.mu.Unlock()
        if _, ok := m.entries[key]; !ok {
                m.order = append(m.order, key)
        }
        m.entries[key] = e
        for m.limit > 0 && len(m.entries) > m.limit {
                old := m.order[0]
                m.order = m.order[1:]
                delete(m.entries, old)
        }
}

func (m *memIndex) Get(key string) (*persist.Entry, bool) {
        m.mu.Lock()
        defer m.mu.Unlock()
        e, ok := m.entries[key]
        return e, ok
}

func (m *memIndex) Delete(key string) {
        m.mu.Lock()
        defer m.mu.Unlock()
        delete(m.entries, key)
}

func (m *memIndex) Len() int {
        m.mu.Lock()
        defer m.mu.Unlock()
        return len(m.entries)
}

func (m *memIndex) Clear() {
        m.mu.Lock()
        defer m.mu.Unlock()
        m.entries = make(map[string]*persist.Entry)
        m.order = nil
}

func (m *memIndex) Keys() []string {
        m.mu.Lock()
        defer m.mu.Unlock()
        out := make([]string, 0, len(m.entries))
        for k := range m.entries {
                out = append(out, k)
        }
        return out
}
