package admin

import (
        "encoding/json"
        "time"

        "github.com/LYH2263/go-imgpipe"
)

// View is a JSON-serializable snapshot for the management UI.
type View struct {
        Time       time.Time    `json:"time"`
        Stats      imgpipe.Stats `json:"stats"`
        CacheKeys  []string     `json:"cache_keys"`
        CacheLen   int          `json:"cache_len"`
        DefaultEnc string       `json:"default_encode"`
}

// Build constructs a View from pipeline state.
func Build(p *imgpipe.Pipeline) View {
        v := View{Time: time.Now().UTC()}
        if p == nil {
                return v
        }
        v.Stats = p.Stats()
        if c := p.Cache(); c != nil {
                v.CacheKeys = c.Keys()
                v.CacheLen = c.Len()
        }
        return v
}

// JSON marshals view with indent.
func JSON(v View) ([]byte, error) {
        return json.MarshalIndent(v, "", "  ")
}
