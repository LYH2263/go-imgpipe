package persist

import (
        "path/filepath"
        "strings"
)

// SafeBlobName ensures key-derived filenames stay within blobs/.
func SafeBlobName(key string) string {
        key = strings.ReplaceAll(key, "..", "")
        key = strings.ReplaceAll(key, "/", "_")
        key = strings.ReplaceAll(key, `\`, "_")
        if key == "" {
                key = "empty"
        }
        if len(key) > 200 {
                key = key[:200]
        }
        return key + ".bin"
}

// BlobPath joins dir/blobs/name.
func BlobPath(dir, name string) string {
        return filepath.Join(dir, "blobs", name)
}
