# 教程：使用 go-modern-ai-stack 构建 Go 服务

一份逐步走通脚手架的指南。学完后你将加一个新服务、写完能通过 race detector 的测试、并学会让 AI agent 遵守你的约定。

## 1. 这是什么脚手架？

`go-modern-ai-stack` 是一个面向 **AI agent 辅助开发** 的 Go 脚手架。它的核心理念：

> 当项目的约定明确、可执行、可验证时，AI agent 产出的代码质量更高。

脚手架通过三种机制实现这个目标：

1. **`AGENTS.md`** —— AI agent 在写代码前读的单一文件，包含项目约定、反模式和默认选择。
2. **一个示例服务** —— `cmd/echo` 和 `internal/echo` 演示了所有约定。AI agent 在生成新代码时可以照着模式匹配。
3. **受约束的工具箱** —— `net/http`、`log/slog`、`database/sql`、无 ORM、无配置库。决策越少 = 错误越少 = 一致性越高。

代价是灵活性下降。你需要承诺：

- 不使用第三方 HTTP 框架（chi、gin、echo、fiber）
- 不使用 ORM（gorm、ent、sqlx-as-ORM）
- 不使用配置库（viper、koanf）
- 不使用日志库（zap、zerolog）
- 不使用依赖注入容器（wire、fx）

如果你的项目真的需要其中任何一个，这个脚手架就不是合适的起点。如果你的项目能接受 stdlib + 测试用 `go-cmp`，这个脚手架将为你节省数周的反复决策成本。

## 2. 前置条件

- **Go 1.23 或更高版本**。运行 `go version` 验证。
- **熟悉 Go 工具链**。你应该能熟练使用 `go build`、`go test`、`go vet`、`gofmt`。
- **一个 AI agent CLI（推荐）**。Codex CLI、Claude Code，或者任何能读项目级指令文件的 agent。没有 AI agent，这个脚手架也能用——只是发挥不了核心价值。

可选但有用：

- **`gh` CLI** —— 用于 GitHub 操作。
- **`make`** —— 用于 v0.2 的 Makefile targets。

## 3. 快速开始

### 3.1 克隆并验证

```bash
git clone https://github.com/gyc567/go-modern-ai-stack.git
cd go-modern-ai-stack
go build ./...
go test -race -count=1 ./...
```

你应该看到：

```
?   	github.com/gyc567/go-modern-ai-stack/cmd/echo	[no test files]
ok  	github.com/gyc567/go-modern-ai-stack/internal/echo	0.5s
```

### 3.2 运行示例服务

```bash
go run ./cmd/echo
```

在另一个终端：

```bash
curl http://localhost:8080/healthz
# ok
curl -X POST -d '你好世界' http://localhost:8080/echo
# {"length":12,"content_type":"","body":"你好世界"}
```

用 `Ctrl-C` 停止服务。服务器会记录 "shutdown signal received" 并干净退出。

### 3.3 刚才发生了什么？

- 服务在 `:8080` 启动（默认；用 `PORT=:9000 go run ./cmd/echo` 覆盖）
- `GET /healthz` 返回 200 和 "ok"
- `POST /echo` 读取 body，按原始字节解析，返回 JSON 信封
- `Ctrl-C` 触发 `SIGINT`，`signal.NotifyContext` 把它转成 context 取消，`srv.Shutdown` 优雅关闭

## 4. 项目布局

```
go-modern-ai-stack/
├── AGENTS.md                  ← AI agent 规则（先读这个）
├── README.md                  ← 人类快速开始
├── LICENSE                    ← MIT
├── docs/
│   ├── tutorial.en.md         ← 本教程（英文）
│   └── tutorial.zh-CN.md      ← 教程（中文）
├── go.mod / go.sum            ← Go 模块
├── .gitignore
├── cmd/
│   └── echo/                  ← 服务二进制入口
│       └── main.go
├── internal/
│   └── echo/                  ← bounded context: echo handlers
│       ├── handler.go
│       └── handler_test.go
└── testdata/                  ← 夹具（当前为空）
```

### 4.1 为什么是这种布局？

- **`cmd/<service>/main.go`** —— 每个服务一个二进制。入口点从目录树上一目了然。
- **`internal/<bounded-context>/`** —— Go 编译器强制 `internal/` 不能被模块外部导入。业务逻辑天然限定在服务内。
- **`pkg/`** —— 默认为空。只有当代码确实被两个或更多服务复用时才往这里放。抵制把它当作"杂物抽屉"的冲动——那会导致纠缠的依赖图。
- **`testdata/`** —— 夹具、golden file、大型测试输入。Go 工具链在编译时跳过这个目录。

### 4.2 命名

- **包名 = 目录名**，小写、单数、不能用下划线。`userservice`，不是 `userService` 也不是 `user_service`。
- **文件名**：小写、snake_case。`user_handler.go`，不是 `UserHandler.go`。
- **避免 stutter**：`userservice.User`，不是 `userservice.UserService`。

## 5. 详解 echo 服务

按这个顺序读示例：`cmd/echo/main.go`，然后 `internal/echo/handler.go`，然后 `internal/echo/handler_test.go`。

### 5.1 `cmd/echo/main.go` —— 二进制结构

每个服务二进制都遵循这个骨架：

```go
func main() {
    if err := run(); err != nil {
        slog.Error("...", "err", err)
        os.Exit(1)
    }
}

func run() error {
    log := slog.New(slog.NewJSONHandler(os.Stdout, ...))
    slog.SetDefault(log)

    // 构建 mux、server 等
    mux := http.NewServeMux()
    mux.Handle("GET /healthz", echo.NewHealthHandler())
    mux.Handle("POST /echo", echo.NewEchoHandler(log))

    srv := &http.Server{
        Addr:         port,
        Handler:      mux,
        ReadTimeout:  defaultReadTimeout,
        WriteTimeout: defaultWriteTimeout,
    }

    // 通过 signal.NotifyContext 实现优雅关闭
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
    defer stop()

    errCh := make(chan error, 1)
    go func() {
        log.InfoContext(ctx, "service listening", "addr", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            errCh <- fmt.Errorf("listen: %w", err)
        }
        close(errCh)
    }()

    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
        log.InfoContext(context.Background(), "shutdown signal received")
    }

    shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownWait)
    defer cancel()

    if err := srv.Shutdown(shutdownCtx); err != nil {
        return fmt.Errorf("shutdown: %w", err)
    }
    return nil
}
```

这个结构强制三条不变量：

1. **`main()` 是唯一调用 `os.Exit` 的地方**。这让二进制可以当作函数测试（理论上可以从测试里调用 `run()`）。
2. **`run()` 返回 error** 而不是 panic 或直接调 `os.Exit`。这分离了"退出逻辑"和"关闭逻辑"。
3. **优雅关闭通过 `signal.NotifyContext` 实现**。关闭超时（默认 10s）是有界的——如果 server 在时间内没排干，`srv.Shutdown` 返回 error。

### 5.2 `internal/echo/handler.go` —— handler 模式

两个 handler：`/healthz`（轻量存活检查）和 `/echo`（实际功能）。

**Handler 构造函数显式接收依赖**：

```go
func NewEchoHandler(log *slog.Logger) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ...
    })
}
```

没有全局变量。logger 通过参数传入。测试里可以传一个丢弃日志的 logger。

**Context 从 request 流入**：

```go
ctx := r.Context()
log.InfoContext(ctx, "echo served", "length", len(body))
```

每行日志都携带 request context，传递 trace ID、超时和取消信号。

**错误用操作上下文包装**：

```go
if err != nil {
    status := http.StatusBadRequest
    var maxBytesErr *http.MaxBytesError
    if errors.As(err, &maxBytesErr) {
        status = http.StatusRequestEntityTooLarge
    }
    respondError(ctx, w, log, status, fmt.Errorf("read body: %w", err))
    return
}
```

`%w`（不是 `%v`）至关重要——它让 `errors.Is` 和 `errors.As` 能解开原始错误。这对以下场景很重要：重试逻辑需要识别 transient 失败、HTTP handler 需要把 error 映射成 status code、日志想暴露底层原因。

**Body 大小有界**：

```go
body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxBodyBytes))
```

没有这个，一个恶意或有 bug 的客户端可以发 10GB body 把你的内存耗尽。`MaxBodyBytes = 1 << 20` 把上限设为 1 MiB。

### 5.3 `internal/echo/handler_test.go` —— 测试模式

三个测试函数：

1. `TestNewHealthHandler` —— 平凡 handler 的单次测试
2. `TestNewEchoHandler` —— echo 功能的表驱动测试
3. `TestNewEchoHandler_TooLarge` —— body 大小限制的边界测试

**表驱动模式**：

```go
cases := []struct {
    name         string
    method       string
    contentType  string
    body         string
    wantStatus   int
    wantInBody   string
    wantEchoBody string
}{
    {name: "text body", method: http.MethodPost, contentType: "text/plain", body: "hello world", ...},
    {name: "empty body", method: http.MethodPost, contentType: "text/plain", body: "", ...},
    {name: "json body", method: http.MethodPost, contentType: "application/json", body: `{"a":1}`, ...},
    {name: "no content type", method: http.MethodPost, contentType: "", body: "raw", ...},
}

for _, tc := range cases {
    t.Run(tc.name, func(t *testing.T) {
        t.Parallel()
        // ... 用 tc ...
    })
}
```

**外层和内层测试都加 `t.Parallel()`**：启用并发测试执行。

**用 `github.com/google/go-cmp` 做深度比较**：

```go
want := echo.EchoResponse{Length: 11, ContentType: "text/plain", Body: "hello world"}
if diff := cmp.Diff(want, got); diff != "" {
    t.Fatalf("mismatch (-want +got):\n%s", diff)
}
```

`cmp.Diff` 产生人类可读的输出，精确显示哪些字段不同。

**测试里丢弃日志**：

```go
log := slog.New(slog.NewTextHandler(io.Discard, nil))
```

不这么做，每次测试都会刷一堆日志到终端。

## 6. 构建你的第一个服务

让我们加一个 `greeter` 服务，响应 `GET /hello/{name}` 返回 `"Hello, {name}!"`。

### 6.1 规划 bounded context

bounded context 是 `greeter`。所有 greeter 代码放在 `internal/greeter/`。如果 greeter 长到需要多个子域（比如 `greeting`、`audience`），就拆开：`internal/greeting/`、`internal/audience/`。

### 6.2 创建 `internal/greeter/handler.go`

```go
// Package greeter 实现 /hello/{name} HTTP handler.
package greeter

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "log/slog"
    "net/http"
)

// Greeting 是 GET /hello/{name} 的 JSON 响应体.
type Greeting struct {
    Name    string `json:"name"`
    Message string `json:"message"`
}

// errEmptyName 是路径值缺失的哨兵 error.
var errEmptyName = errors.New("empty name")

// NewGreetingHandler 返回一个响应个性化问候语的 handler.
func NewGreetingHandler(log *slog.Logger) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        name := r.PathValue("name")
        if name == "" {
            respondError(r.Context(), w, log, http.StatusBadRequest,
                fmt.Errorf("name required: %w", errEmptyName))
            return
        }

        msg := fmt.Sprintf("Hello, %s!", name)
        log.InfoContext(r.Context(), "greeting served", "name", name)

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        if err := json.NewEncoder(w).Encode(Greeting{Name: name, Message: msg}); err != nil {
            log.ErrorContext(r.Context(), "encode response", "err", fmt.Errorf("encode: %w", err))
            return
        }
    })
}

func respondError(ctx context.Context, w http.ResponseWriter, log *slog.Logger, code int, err error) {
    log.WarnContext(ctx, "handler error", "code", code, "err", err)
    http.Error(w, http.StatusText(code), code)
}
```

### 6.3 创建 `internal/greeter/handler_test.go`

```go
package greeter_test

import (
    "encoding/json"
    "io"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/google/go-cmp/cmp"

    "github.com/gyc567/go-modern-ai-stack/internal/greeter"
)

func TestNewGreetingHandler(t *testing.T) {
    t.Parallel()

    cases := []struct {
        name       string
        path       string
        wantStatus int
        want       greeter.Greeting
    }{
        {
            name:       "normal",
            path:       "/hello/Alice",
            wantStatus: http.StatusOK,
            want:       greeter.Greeting{Name: "Alice", Message: "Hello, Alice!"},
        },
        {
            name:       "encoded space",
            path:       "/hello/Bob%20Smith",
            wantStatus: http.StatusOK,
            want:       greeter.Greeting{Name: "Bob Smith", Message: "Hello, Bob Smith!"},
        },
        {
            name:       "empty path segment",
            path:       "/hello/",
            wantStatus: http.StatusBadRequest,
        },
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()

            log := slog.New(slog.NewTextHandler(io.Discard, nil))
            rec := httptest.NewRecorder()
            req := httptest.NewRequest(http.MethodGet, tc.path, nil)

            greeter.NewGreetingHandler(log).ServeHTTP(rec, req)

            if rec.Code != tc.wantStatus {
                t.Fatalf("status: got %d, want %d; body=%s",
                    rec.Code, tc.wantStatus, rec.Body.String())
            }

            if tc.wantStatus != http.StatusOK {
                return
            }

            var got greeter.Greeting
            if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
                t.Fatalf("decode response: %v", err)
            }
            if diff := cmp.Diff(tc.want, got); diff != "" {
                t.Fatalf("response mismatch (-want +got):\n%s", diff)
            }
        })
    }
}
```

### 6.4 验证

```bash
go build ./...
go test -race -count=1 ./...
go vet ./...
gofmt -l .
```

这四条命令都应该过。如果没过，看下面的故障排查。

### 6.5 接入二进制

创建 `cmd/greeter/main.go` —— 复制 `cmd/echo/main.go` 然后改路由注册：

```go
mux.Handle("GET /hello/{name}", greeter.NewGreetingHandler(log))
```

编译并运行：

```bash
go run ./cmd/greeter
# 在另一个终端：
curl http://localhost:8080/hello/World
# {"name":"World","message":"Hello, World!"}
```

## 7. 配合 AI agent 工作

这是脚手架的核心价值。

### 7.1 让 agent 读 AGENTS.md

大多数现代 AI agent（Codex CLI、Claude Code、Cursor、Aider、Cline）都会读项目级指令文件。`AGENTS.md` 这个约定被广泛支持。

打开你的 agent 并提示：

> "在 `internal/todo/` 加一个新服务，在内存里管理 TODO 列表。服务要暴露 `POST /todos`（创建）、`GET /todos`（列表）、`DELETE /todos/{id}`（删除）。严格遵守 AGENTS.md。"

### 7.2 验证 agent 的输出

agent 写完代码后，跑验证循环：

```bash
go build ./...
go test -race -count=1 ./...
go vet ./...
gofmt -l .
```

如果任何命令失败：

- **读 error**。编译器和 linter 是确定性的；error 告诉你确切哪里错了。
- **识别 agent 漏掉的规则**。交叉对照 AGENTS.md。
- **修或退回**。你可以自己修（更快、树立榜样）或让 agent 修（教 agent）。

### 7.3 AI agent 的常见错误

即使有 AGENTS.md，agent 还是会：

1. **忘记包装 error**：`return err` 而不是 `return fmt.Errorf("op: %w", err)`
2. **handler 里丢掉 ctx**：用 `log.Info` 而不是 `log.InfoContext(ctx, ...)`
3. **测试里用 `time.Sleep`**：应该用带 deadline 的 polling
4. **引入配置库**：viper、koanf、envconfig
5. **业务路径用 `panic`**：应该返回 error
6. **忘记 `t.Parallel()`**：外层或内层测试
7. **用 `reflect.DeepEqual`**：应该用 go-cmp 的 `cmp.Diff`
8. **创建 `pkg/utils/`**：应该放在 `internal/<bounded-context>/`

当你发现**重复**的错误时，**把规则加到 AGENTS.md 里，配 BAD/GOOD 示例**。这能把一次性纠正变成对未来 session 的永久指导。

### 7.4 构建 agent 友好的反馈循环

有效的模式：

```
1. 给 agent 一个明确的任务提示
2. 跑验证命令
3. 如果有失败：
   a. 识别漏掉了哪条 AGENTS.md 规则
   b. 让 agent 修，引用那条规则的 BAD/GOOD 示例
   c. 如果 agent 重复犯同样错误，升级到把规则加进 AGENTS.md（更明确、再加一个 BAD/GOOD）
4. 类似代码成功生成 2-3 次后，agent 已经内化了约定
```

这比自己修慢，但能建立持久的 agent 行为。

## 8. 约定深入

本节展开 AGENTS.md，补充理由和常见陷阱。

### 8.1 Error：永远包装

```go
// BAD: 调用方不知道哪里失败
if err != nil {
    return err
}

// GOOD: 调用方能看到操作上下文
if err != nil {
    return fmt.Errorf("read %s: %w", path, err)
}
```

`%w`（不是 `%v`）至关重要——它让 `errors.Is` 和 `errors.As` 能解开原始错误。这对重试逻辑、HTTP error 到 status code 映射、日志暴露底层原因都很重要。

### 8.2 Context：第一参数

```go
// BAD: 调用方不能取消
func Load() (Config, error) { ... }

// GOOD: 调用方可以取消、设 deadline、附加值
func Load(ctx context.Context) (Config, error) { ... }
```

在 handler 里，context 是 `r.Context()`。永远不要在 handler 里用 `context.Background()` 创建新 context——你会丢掉 request 的 deadline 和取消。

### 8.3 Goroutine：errgroup

```go
// BAD: 静默失败，不取消
go func() {
    work()
}()
go func() {
    moreWork()
}()

// GOOD: 错误传播，第一个失败取消其余
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { return work(ctx) })
g.Go(func() error { return moreWork(ctx) })
if err := g.Wait(); err != nil {
    return err
}
```

当你有多个应该一起失败的 goroutine 时用 `errgroup`。不要用裸 channel 做取消信号。

### 8.4 HTTP：stdlib ServeMux

```go
// BAD: 第三方依赖，其他 Go 开发者不熟悉的 API
r := chi.NewRouter()

// GOOD: stdlib、显式、对大多数服务足够
mux := http.NewServeMux()
mux.HandleFunc("GET /users/{id}", h.GetUser)
mux.HandleFunc("POST /users", h.CreateUser)
```

Go 1.22+ 在 `http.ServeMux` 里直接支持方法模式（`"GET /path"`）和路径参数（`{id}`）。这对 ~90% 的服务足够。剩下的 10% 需要复杂路由、中间件链、router 级别的优雅抽象——它们是例外——它们应该是重新审视这条约定的触发条件。

### 8.5 Logging：带 context 的 slog

```go
// BAD: 没有 context 传播，难以追踪
log.Println("hello")

// GOOD: request 级、结构化、可 grep
slog.InfoContext(ctx, "hello", "user_id", id, "request_id", reqID)
```

如果你用 OpenTelemetry，`slog.InfoContext` 会自动从 context 提取 trace ID。即使没有 OTel，每行日志都携带 request 的生命周期——调试"这个 request 发生了什么？"时无价。

### 8.6 Test：parallel + race

```bash
go test -race -count=1 ./...
```

三个 flag 重要：

- **`-race`**：启用 race 检测器。捕捉数据竞争，否则表现为生产环境的 flaky 行为。
- **`-count=1`**：强制不缓存测试。CI 里重要；本地可以省。
- **`./...`**：跑模块内所有测试。

`t.Parallel()` 让单个测试并发执行。结合 `-race`，能捕捉只在并发执行下出现的同步 bug。

## 9. 何时偏离约定

约定是**默认**，不是法律。在以下情况打破：

1. **某条规则对你的领域不适用**。比如：你的服务是 WebSocket 重度的实时 API，`net/http.ServeMux` 的建议可能不适合。
2. **三个或更多不同的 bounded context 撞同一堵墙**。AGENTS.md 里记录的"3+"规则。
3. **你在 AGENTS.md 里记录了偏离**。这防止下一个贡献者（人或 AI）把你的工作撤销。

示例 diff：

```diff
- ### 11. Database
- `database/sql` with `pgx` (Postgres) 或 `lib/pq`. No ORM.
+ ### 11. Database
+ `database/sql` with `pgx` (Postgres) 或 `lib/pq`. **例外**：`internal/analytics/` 用 `sqlc`（2026-08-26 记录：30+ 带 join 的读查询）。
```

例外可见、有日期、有范围。

## 10. 故障排查

### "build failed: undefined: ..."

你忘了 import 包，或 import 路径错了。看 `go.mod` 找正确的模块路径。

### "test failed: ... got X, want Y"

仔细读失败的断言。如果是结构差异，`cmp.Diff` 会精确显示哪些字段不同。如果是 status code，handler 可能返回了错的 code（对照 AGENTS.md 的 error 约定检查）。

### "go vet: ... loop variable captured by func"

你按引用捕获了 `for` 循环变量。Go 1.22+ 的循环变量是 per-iteration 的，所以这很少见。在更老的 Go 里，通过参数传变量修复：

```go
for _, p := range paths {
    p := p // 仅 Go <1.22
    g.Go(func() error {
        return process(ctx, p)
    })
}
```

### "agent 忽略了 AGENTS.md 第 X 条"

三种选择：

1. **改写规则** —— 更明确。再加一个 BAD/GOOD 示例。
2. **在 prompt 里明确引用规则** —— "按 AGENTS.md 第 5 条，这个 handler 必须用 consumer-side 接口。"
3. **退回并迭代** —— 不要接受代码；让 agent 重做。

## 11. 下一步

这是 v0.1。未来版本：

- **v0.2**：Makefile、golangci-lint 配置、面向 AI 反模式的 audit 脚本、pre-commit hook。
- **v0.3**：通过 `tools/tools.go` 锁定 dev 工具版本、覆盖率门槛。

在那之前，工作流是：写代码 → 跑验证 → 看到重复错误时更新 AGENTS.md。

更多设计哲学见 README 和 AGENTS.md。
