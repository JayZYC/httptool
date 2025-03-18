# httptool

httptool 是一个简单易用的 HTTP 客户端工具包，提供了丰富的功能和灵活的配置选项。该工具包基于 Go 标准库的 `net/http` 包构建，并添加了更多实用的功能。

## 特性

- 支持常用的 HTTP 方法（GET、POST、PUT、DELETE 等）
- 灵活的配置选项系统
- 内置的日志记录功能
- 慢请求监控
- 自定义 HTTP 客户端支持
- 上下文（Context）支持
- 超时控制
- 请求头管理

## 安装

```bash
go get github.com/yourusername/httptool
```

## 快速开始

### 基本 GET 请求

```go
import "github.com/yourusername/httptool"

func main() {
    ctx := context.Background()
    statusCode, body, err := httptool.Get(ctx, "https://api.example.com/data")
    if err != nil {
        // 处理错误
        return
    }
    // 处理响应
}
```

### 带选项的 POST 请求

```go
data := []byte(`{
    "name": "张三",
    "age": 25
}`)
options := []httptool.Option{
    httptool.WithTimeout(10 * time.Second),
    httptool.WithHeaders(map[string]string{
        "Authorization": "Bearer token123",
    }),
    httptool.WithSlowThreshold(100 * time.Millisecond),
}
statusCode, body, err := httptool.Post(ctx, "https://api.example.com/users", data, options...)
```

## 配置选项

httptool 提供了多种配置选项，可以根据需要组合使用：

### WithTimeout
设置请求超时时间：
```go
httptool.WithTimeout(10 * time.Second)
```

### WithHeaders
设置请求头：
```go
httptool.WithHeaders(map[string]string{
    "Authorization": "Bearer token123",
    "Content-Type": "application/json",
})
```

### WithSlowThreshold
设置慢请求阈值：
```go
httptool.WithSlowThreshold(100 * time.Millisecond)
```

### WithLogger
设置自定义日志记录器：
```go
customLogger := httptool.New(log.New(os.Stdout, "", log.LstdFlags), httptool.Config{
    LogLevel: httptool.Debug,
    Colorful: true,
})
httptool.WithLogger(customLogger)
```

### WithContext
设置请求上下文：
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
httptool.WithContext(ctx)
```

## 自定义 HTTP 客户端

可以创建自定义的 HTTP 客户端并设置为全局客户端：

```go
customClient := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 50,
        IdleConnTimeout:     90 * time.Second,
    },
}
httptool.SetHttpClient(customClient)
```

## 日志功能

httptool 提供了内置的日志记录功能，支持不同的日志级别和彩色输出：

```go
logger := httptool.New(log.New(os.Stdout, "", log.LstdFlags), httptool.Config{
    LogLevel: httptool.Debug,
    Colorful: true,
})
```

支持的日志级别：
- Debug
- Info
- Warn
- Error

## 错误处理

httptool 会返回以下信息：
- statusCode: HTTP 状态码
- body: 响应体
- err: 错误信息

建议总是检查错误：
```go
statusCode, body, err := httptool.Get(ctx, url)
if err != nil {
    // 处理错误
    return
}
```

## 最佳实践

1. 总是使用上下文来控制请求的生命周期
2. 设置适当的超时时间
3. 使用慢请求阈值监控性能
4. 在生产环境中使用自定义日志记录器
5. 根据需要配置自定义 HTTP 客户端

## 示例

更多使用示例请参考 [example_test.go](example_test.go)

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License 