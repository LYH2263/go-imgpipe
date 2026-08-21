package main

import (
        "encoding/base64"
        "encoding/json"
        "flag"
        "fmt"
        "io"
        "log"
        "net/http"
        "os"
        "strconv"
        "time"

        "github.com/LYH2263/go-imgpipe"
        "github.com/LYH2263/go-imgpipe/internal/admin"
        "github.com/LYH2263/go-imgpipe/internal/encode"
        "github.com/LYH2263/go-imgpipe/internal/jobutil"
)

func main() {
        addr := flag.String("addr", ":8111", "listen address")
        cache := flag.String("cache", "./data/cache", "disk cache directory")
        web := flag.String("web", "web", "static web directory")
        flag.Parse()
        if err := os.MkdirAll(*cache, 0o755); err != nil {
                log.Fatal(err)
        }
        pipe, err := imgpipe.New(imgpipe.Options{
                CacheDir:           *cache,
                StripExifByDefault: false,
                DefaultEncode:      imgpipe.EncodeJPEG,
                DefaultQuality:     85,
        })
        if err != nil {
                log.Fatal(err)
        }
        defer pipe.Close()

        mux := http.NewServeMux()
        mux.Handle("/", http.FileServer(http.Dir(*web)))
        mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
                writeJSON(w, map[string]any{"ok": true, "time": time.Now().UTC()})
        })
        mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
                writeJSON(w, admin.Build(pipe))
        })
        mux.HandleFunc("/api/cache/keys", func(w http.ResponseWriter, r *http.Request) {
                c := pipe.Cache()
                keys := []string{}
                if c != nil {
                        keys = c.Keys()
                }
                writeJSON(w, map[string]any{"keys": keys, "len": len(keys)})
        })
        mux.HandleFunc("/api/transform", handleTransform(pipe))
        mux.HandleFunc("/api/preview-key", handlePreviewKey(pipe))

        fmt.Println("imgd listening on", *addr, "cache", *cache)
        log.Fatal(http.ListenAndServe(*addr, mux))
}

func handleTransform(pipe *imgpipe.Pipeline) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
                if r.Method != http.MethodPost {
                        http.Error(w, "POST only", 405)
                        return
                }
                var body []byte
                var err error
                if err = r.ParseMultipartForm(16 << 20); err == nil {
                        file, _, ferr := r.FormFile("file")
                        if ferr != nil {
                                http.Error(w, "file required", 400)
                                return
                        }
                        defer file.Close()
                        body, err = io.ReadAll(io.LimitReader(file, 16<<20))
                } else {
                        body, err = io.ReadAll(io.LimitReader(r.Body, 16<<20))
                }
                if err != nil {
                        http.Error(w, err.Error(), 400)
                        return
                }
                runJob(w, r, pipe, body)
        }
}

func runJob(w http.ResponseWriter, r *http.Request, pipe *imgpipe.Pipeline, raw []byte) {
        transforms, err := jobutil.ParseTransformList(r.URL.Query().Get("t"))
        if err != nil {
                http.Error(w, err.Error(), 400)
                return
        }
        format := r.URL.Query().Get("format")
        if format == "" {
                format = "jpeg"
        }
        q, _ := strconv.Atoi(r.URL.Query().Get("q"))
        if q == 0 {
                q = 85
        }
        skip := r.URL.Query().Get("skip_cache") == "1"
        strip := r.URL.Query().Get("strip_exif") == "1"
        job := imgpipe.Job{
                Raw:        raw,
                Transforms: transforms,
                Encode:     imgpipe.EncodeOptions{Format: imgpipe.EncodeFormat(format), Quality: q},
                SkipCache:  skip,
                StripExif:  strip,
        }
        res, err := pipe.RunContext(r.Context(), job)
        if err != nil {
                http.Error(w, err.Error(), 400)
                return
        }
        if r.URL.Query().Get("meta") == "1" {
                writeJSON(w, map[string]any{
                        "cache_key":   res.CacheKey,
                        "cache_hit":   res.CacheHit,
                        "width":       res.Width,
                        "height":      res.Height,
                        "format":      res.Format,
                        "bytes":       len(res.Bytes),
                        "duration_ms": res.Duration.Milliseconds(),
                        "preview_b64": base64.StdEncoding.EncodeToString(res.Bytes),
                })
                return
        }
        w.Header().Set("Content-Type", encode.ContentType(res.Format))
        w.Header().Set("X-Cache-Key", res.CacheKey)
        w.Header().Set("X-Cache-Hit", strconv.FormatBool(res.CacheHit))
        w.Header().Set("X-Image-Width", strconv.Itoa(res.Width))
        w.Header().Set("X-Image-Height", strconv.Itoa(res.Height))
        _, _ = w.Write(res.Bytes)
}

func handlePreviewKey(pipe *imgpipe.Pipeline) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
                body, err := io.ReadAll(io.LimitReader(r.Body, 16<<20))
                if err != nil {
                        http.Error(w, err.Error(), 400)
                        return
                }
                transforms, err := jobutil.ParseTransformList(r.URL.Query().Get("t"))
                if err != nil {
                        http.Error(w, err.Error(), 400)
                        return
                }
                format := r.URL.Query().Get("format")
                if format == "" {
                        format = "jpeg"
                }
                q, _ := strconv.Atoi(r.URL.Query().Get("q"))
                job := imgpipe.Job{
                        Raw:        body,
                        Transforms: transforms,
                        Encode:     imgpipe.EncodeOptions{Format: imgpipe.EncodeFormat(format), Quality: q},
                        StripExif:  r.URL.Query().Get("strip_exif") == "1",
                }
                writeJSON(w, map[string]string{"cache_key": pipe.PreviewKey(job)})
        }
}

func writeJSON(w http.ResponseWriter, v any) {
        w.Header().Set("Content-Type", "application/json")
        enc := json.NewEncoder(w)
        enc.SetIndent("", "  ")
        _ = enc.Encode(v)
}
