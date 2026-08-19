# logkv — 嵌入式日志结构 KV 存储

logkv 是一个 Bitcask 风格的嵌入式键值存储库，使用 bbolt 作为持久化后端，在内存中维护有序索引，并提供 TTL 过期清扫、压缩（compaction）、备份/恢复与导出能力。它通过 HTTP API 暴露上述能力，适用于需要本地持久化、重启可恢复的轻量级存储场景。

## 主要能力

- 有序键值存储，支持按前缀列举与按范围扫描。
- 每条记录可设置 TTL，过期后由后台清扫器自动回收。
- 支持压缩以回收 tombstone 与过期记录占用的空间。
- 支持把全量 live 记录备份为二进制快照、从快照恢复、以及带 context 取消语义的导出。
- 提供 20 个 HTTP API 操作（PUT/GET/DELETE/TTL/批量/范围/统计/备份恢复等）。

## 标准构建、运行与测试命令

以下命令均假设当前目录为项目根（包含 `go.mod`）：

- 编译：`go build ./...`
- 启动（自检模式，执行后自行退出）：`go run . --smoke-test`
- 启动（常驻 HTTP 服务，监听 `:8080`）：`go run . --addr :8080`
- 测试：`go test ./...`
- 静态检查：`go vet ./...`

本项目为纯 Go 项目，使用 Go 1.26.3（`GOTOOLCHAIN=local`，`go.mod` 语言版本 `go 1.26.3`）。依赖通过 Go module mode 下载；仓库不提交 `vendor/`，构建与测试不使用 `-mod` 参数。

## Benzhi Docker 构建

`build_benzhi_docker.sh` 使用固定的 `benzhi.Dockerfile` 构建镜像，不依赖本机 vendor 目录，依赖在容器内下载：

```bash
# 默认 linux/amd64
./build_benzhi_docker.sh my-logkv
# 指定平台
./build_benzhi_docker.sh my-logkv linux/arm64
```

镜像基于 `docker.m.daocloud.io/library/golang:1.26.3-bookworm`，启动后进入 bash 便于交互排查。运行时可直接执行：

```bash
docker run -it my-logkv:latest
go build ./...
go test ./...
```

## 备注

- 数据持久化使用 bbolt 单文件数据库，路径由 `--db` 指定（默认 `logkv.db`）。
- `--smoke-test` 为无外部依赖的自检，执行端到端读写、范围扫描、备份恢复、导出取消与压缩后退出，exit code 0 表示通过。
