# go-imgpipe

图片变换流水线：载入字节 → 解码（JPEG/PNG/GIF）→ 变换链（缩放 / 裁剪 / 水印叠图 / 旋转 / 翻转）→ 编码输出 → 内容寻址缓存键（内存 + 可选磁盘）。

## 模块

```text
github.com/LYH2263/go-imgpipe
```

主类型：`Pipeline`、`Frame`、`Job`。

## 构建与测试

```text
go build ./...
go test ./... -count=1
```

## 管理页

```text
go run ./cmd/imgd -addr :8111 -web web -cache ./data/cache
```

浏览器打开 `http://127.0.0.1:8111/`，上传图片试跑变换并查看缓存键。

## 包结构

- `pipeline.go` / `frame.go` / `cache.go` / `options.go` / `close.go`
- `internal/decode` · `encode` · `xform` · `exifstrip` · `cachekey` · `persist` · `pool` · `validate` · `metrics`
