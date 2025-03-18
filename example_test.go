package httptool

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"
)

// ExampleGet 展示GET请求的使用方法
func ExampleGet() {
	// 创建上下文
	ctx := context.Background()

	// 基本GET请求
	statusCode, body, err := Get(ctx, "https://api.example.com/data")
	if err != nil {
		// 处理错误
		return
	}
	_ = statusCode
	_ = body

	// 带自定义选项的GET请求
	options := []Option{
		WithTimeout(10 * time.Second), // 设置超时
		WithHeaders(map[string]string{ // 设置请求头
			"Authorization":   "Bearer token123",
			"X-Custom-Header": "value",
		}),
		WithSlowThreshold(100 * time.Millisecond), // 设置慢请求阈值
	}
	statusCode, body, err = Get(ctx, "https://api.example.com/data", options...)
}

// ExamplePost 展示POST请求的使用方法
func ExamplePost() {
	ctx := context.Background()

	// 准备POST数据
	data := []byte(`{
		"name": "张三",
		"age": 25,
		"email": "zhangsan@example.com"
	}`)

	// 基本POST请求
	statusCode, body, err := Post(ctx, "https://api.example.com/users", data)
	if err != nil {
		// 处理错误
		return
	}
	_ = statusCode
	_ = body

	// 带自定义选项的POST请求
	options := []Option{
		WithTimeout(10 * time.Second),
		WithHeaders(map[string]string{
			"Authorization": "Bearer token123",
		}),
		WithSlowThreshold(100 * time.Millisecond),
	}
	statusCode, body, err = Post(ctx, "https://api.example.com/users", data, options...)
}

// ExampleRequest 展示通用Request方法的使用
func ExampleRequest() {
	ctx := context.Background()

	// PUT请求示例
	putData := []byte(`{
		"name": "李四",
		"age": 30
	}`)
	options := []Option{
		WithContext(ctx),
		WithData(putData),
		WithHeaders(map[string]string{
			"Authorization": "Bearer token123",
		}),
		WithTimeout(10 * time.Second),
	}
	statusCode, body, err := Request("PUT", "https://api.example.com/users/1", options...)
	if err != nil {
		// 处理错误
		return
	}
	_ = statusCode
	_ = body

	// DELETE请求示例
	statusCode, body, err = Request("DELETE", "https://api.example.com/users/1", WithContext(ctx))
}

// ExampleCustomClient 展示自定义HTTP客户端的使用
func ExampleCustomClient() {
	// 创建自定义的HTTP客户端
	customClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 50,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// 设置全局HTTP客户端
	SetHttpClient(customClient)

	// 使用自定义客户端发送请求
	ctx := context.Background()
	statusCode, body, err := Get(ctx, "https://api.example.com/data")
	if err != nil {
		// 处理错误
		return
	}
	_ = statusCode
	_ = body
}

// ExampleWithLogger 展示自定义日志记录器的使用
func ExampleWithLogger() {
	ctx := context.Background()

	// 创建自定义日志记录器
	customLogger := New(log.New(os.Stdout, "", log.LstdFlags), Config{
		LogLevel: Debug,
		Colorful: true,
	})

	// 使用自定义日志记录器发送请求
	options := []Option{
		WithContext(ctx),
		WithLogger(customLogger),
		WithSlowThreshold(100 * time.Millisecond),
	}
	statusCode, body, err := Get(ctx, "https://api.example.com/data", options...)
	if err != nil {
		// 处理错误
		return
	}
	_ = statusCode
	_ = body
}

// ExampleWithContext 展示使用带超时的上下文
func ExampleWithContext() {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 使用带超时的上下文发送请求
	statusCode, body, err := Get(ctx, "https://api.example.com/data")
	if err != nil {
		// 处理错误
		return
	}
	_ = statusCode
	_ = body
}

// ExampleWithSlowThreshold 展示慢请求监控
func ExampleWithSlowThreshold() {
	ctx := context.Background()

	// 设置慢请求阈值为100毫秒
	options := []Option{
		WithContext(ctx),
		WithSlowThreshold(100 * time.Millisecond),
	}

	// 发送请求，如果响应时间超过阈值，会记录警告日志
	statusCode, body, err := Get(ctx, "https://api.example.com/data", options...)
	if err != nil {
		// 处理错误
		return
	}
	_ = statusCode
	_ = body
}
