package imgpipe

// String returns a compact stats summary.
func (s Stats) String() string {
        closed := "open"
        if s.Closed {
                closed = "closed"
        }
        return sprintfStats(s, closed)
}

func sprintfStats(s Stats, closed string) string {
        // keep fmt usage localized
        return formatStats(s.Runs, s.CacheHits, s.CacheMiss, s.DecodeFail, s.EncodeFail, s.BytesIn, s.BytesOut, closed)
}
