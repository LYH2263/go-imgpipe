package imgpipe

import "fmt"

func formatStats(runs, hits, miss, dfail, efail, in, out int64, closed string) string {
        return fmt.Sprintf("runs=%d hits=%d miss=%d decode_fail=%d encode_fail=%d in=%d out=%d %s",
                runs, hits, miss, dfail, efail, in, out, closed)
}
