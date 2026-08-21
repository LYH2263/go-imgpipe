package persist

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Entry is one cached frame blob.
type Entry struct {
	Key    string
	W      int
	H      int
	Stride int
	Format string
	Pixels []byte
	Raw    []byte
}

// Options configures on-disk cache.
type Options struct {
	Dir         string
	JournalName string
	SyncOnWrite bool
	LimitBytes  int64
}

// Store is a content-addressed disk cache with a journal index.
type Store struct {
	opts    Options
	mu      sync.Mutex
	index   map[string]string // key → relative filename
	journal *os.File
	closed  bool
	bytes   int64
}

type journalRec struct {
	Op   string `json:"op"`
	Key  string `json:"key"`
	File string `json:"file,omitempty"`
	W    int    `json:"w,omitempty"`
	H    int    `json:"h,omitempty"`
}

// Open creates or loads a disk cache directory.
func Open(opts Options) (*Store, error) {
	if opts.Dir == "" {
		return nil, errors.New("persist: empty dir")
	}
	if opts.JournalName == "" {
		opts.JournalName = "cache.journal"
	}
	if err := os.MkdirAll(opts.Dir, 0o755); err != nil {
		return nil, err
	}
	blobs := filepath.Join(opts.Dir, "blobs")
	if err := os.MkdirAll(blobs, 0o755); err != nil {
		return nil, err
	}
	s := &Store{
		opts:  opts,
		index: make(map[string]string),
	}
	jpath := filepath.Join(opts.Dir, opts.JournalName)
	if err := s.replay(jpath); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(jpath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	s.journal = f
	return s, nil
}

func (s *Store) replay(jpath string) error {
	data, err := os.ReadFile(jpath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	// journal is newline-delimited JSON
	start := 0
	for i := 0; i <= len(data); i++ {
		if i == len(data) || data[i] == '\n' {
			line := data[start:i]
			start = i + 1
			if len(line) == 0 {
				continue
			}
			var rec journalRec
			if err := json.Unmarshal(line, &rec); err != nil {
				continue
			}
			switch rec.Op {
			case "put":
				s.index[rec.Key] = rec.File
			case "del":
				delete(s.index, rec.Key)
			}
		}
	}
	return nil
}

// Put writes entry to a temp file, Syncs, renames, then appends journal.
// If rename fails the memory/journal index is NOT updated.
func (s *Store) Put(e *Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.New("persist: closed")
	}
	name := e.Key + ".bin"
	final := filepath.Join(s.opts.Dir, "blobs", name)
	tmp := final + ".tmp"
	if err := writeEntryFile(tmp, e, s.opts.SyncOnWrite); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, final); err != nil {
		_ = os.Remove(tmp)
		return err // do not update index
	}
	rec := journalRec{Op: "put", Key: e.Key, File: name, W: e.W, H: e.H}
	if err := s.appendJournal(rec); err != nil {
		// best-effort: file exists but journal failed — still index for this process
		s.index[e.Key] = name
		return err
	}
	s.index[e.Key] = name
	s.bytes += int64(len(e.Pixels) + len(e.Raw))
	return nil
}

func writeEntryFile(path string, e *Entry, syncOn bool) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	// simple binary: magic + dims + pixels len + pixels + raw len + raw
	if _, err := f.Write([]byte("IPB1")); err != nil {
		return err
	}
	hdr := make([]byte, 16)
	binary.LittleEndian.PutUint32(hdr[0:], uint32(e.W))
	binary.LittleEndian.PutUint32(hdr[4:], uint32(e.H))
	binary.LittleEndian.PutUint32(hdr[8:], uint32(e.Stride))
	binary.LittleEndian.PutUint32(hdr[12:], uint32(len(e.Format)))
	if _, err := f.Write(hdr); err != nil {
		return err
	}
	if _, err := f.Write([]byte(e.Format)); err != nil {
		return err
	}
	if err := writeBlob(f, e.Pixels); err != nil {
		return err
	}
	if err := writeBlob(f, e.Raw); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}
	if syncOn {
		if err := f.Sync(); err != nil {
			return err
		}
	}
	return nil
}

func writeBlob(f *os.File, b []byte) error {
	var ln [4]byte
	binary.LittleEndian.PutUint32(ln[:], uint32(len(b)))
	if _, err := f.Write(ln[:]); err != nil {
		return err
	}
	if len(b) > 0 {
		_, err := f.Write(b)
		return err
	}
	return nil
}

func (s *Store) appendJournal(rec journalRec) error {
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if _, err := s.journal.Write(b); err != nil {
		return err
	}
	if s.opts.SyncOnWrite {
		return s.journal.Sync()
	}
	return nil
}

// Get loads an entry by key.
func (s *Store) Get(key string) (*Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, errors.New("persist: closed")
	}
	name, ok := s.index[key]
	if !ok {
		return nil, fmt.Errorf("persist: miss %s", key)
	}
	path := filepath.Join(s.opts.Dir, "blobs", name)
	return readEntryFile(path, key)
}

func readEntryFile(path, key string) (*Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) < 4+16 || string(data[:4]) != "IPB1" {
		return nil, errors.New("persist: bad magic")
	}
	off := 4
	w := int(binary.LittleEndian.Uint32(data[off:]))
	h := int(binary.LittleEndian.Uint32(data[off+4:]))
	stride := int(binary.LittleEndian.Uint32(data[off+8:]))
	flen := int(binary.LittleEndian.Uint32(data[off+12:]))
	off += 16
	if off+flen > len(data) {
		return nil, errors.New("persist: truncate format")
	}
	format := string(data[off : off+flen])
	off += flen
	pixels, off, err := readBlob(data, off)
	if err != nil {
		return nil, err
	}
	raw, _, err := readBlob(data, off)
	if err != nil {
		return nil, err
	}
	return &Entry{Key: key, W: w, H: h, Stride: stride, Format: format, Pixels: pixels, Raw: raw}, nil
}

func readBlob(data []byte, off int) ([]byte, int, error) {
	if off+4 > len(data) {
		return nil, off, errors.New("persist: truncate len")
	}
	n := int(binary.LittleEndian.Uint32(data[off:]))
	off += 4
	if off+n > len(data) {
		return nil, off, errors.New("persist: truncate blob")
	}
	b := make([]byte, n)
	copy(b, data[off:off+n])
	return b, off + n, nil
}

// Has reports index presence.
func (s *Store) Has(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.index[key]
	return ok
}

// Delete removes key from index and blob file.
func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	name, ok := s.index[key]
	if !ok {
		return nil
	}
	delete(s.index, key)
	_ = s.appendJournal(journalRec{Op: "del", Key: key})
	return os.Remove(filepath.Join(s.opts.Dir, "blobs", name))
}

// Sync flushes the journal to stable storage.
func (s *Store) Sync() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.journal == nil {
		return nil
	}
	return s.journal.Sync()
}

// Close syncs then closes the journal.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	var err error
	if s.journal != nil {
		err = s.journal.Sync()
		if cerr := s.journal.Close(); cerr != nil && err == nil {
			err = cerr
		}
		s.journal = nil
	}
	return err
}

// Keys returns indexed keys.
func (s *Store) Keys() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, 0, len(s.index))
	for k := range s.index {
		out = append(out, k)
	}
	return out
}
