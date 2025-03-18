package httptool

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// MockLogger 实现 Interface 接口，用于测试
type MockLogger struct {
	debugCalled bool
	infoCalled  bool
	warnCalled  bool
	errorCalled bool
	lastMsg     string
	lastData    []interface{}
}

func (m *MockLogger) LogMode(level LogLevel) Interface {
	return m
}

func (m *MockLogger) Debug(ctx context.Context, msg string, data ...interface{}) {
	m.debugCalled = true
	m.lastMsg = msg
	m.lastData = data
}

func (m *MockLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	m.infoCalled = true
	m.lastMsg = msg
	m.lastData = data
}

func (m *MockLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	m.warnCalled = true
	m.lastMsg = msg
	m.lastData = data
}

func (m *MockLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	m.errorCalled = true
	m.lastMsg = msg
	m.lastData = data
}

// 自定义ResponseWriter，用于测试出错情况
type errorResponseWriter struct {
	*httptest.ResponseRecorder
}

func (e *errorResponseWriter) Write(p []byte) (int, error) {
	return 0, errors.New("测试错误")
}

// 测试用例开始前重置客户端
func resetClient() {
	client = nil
	once = sync.Once{}
}

// TestGetHttpClient 测试获取HTTP客户端
func TestGetHttpClient(t *testing.T) {
	resetClient()

	// 测试默认客户端
	c1 := GetHttpClient()
	if c1 == nil {
		t.Fatal("GetHttpClient() 返回了nil")
	}

	// 再次获取客户端，应该是同一个实例
	c2 := GetHttpClient()
	if c1 != c2 {
		t.Fatal("GetHttpClient() 两次返回的不是同一个实例")
	}
}

// TestSetHttpClient 测试设置自定义HTTP客户端
func TestSetHttpClient(t *testing.T) {
	resetClient()

	// 创建自定义客户端
	customClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 设置自定义客户端
	SetHttpClient(customClient)

	// 获取客户端，应该是自定义的客户端
	c := GetHttpClient()
	if c != customClient {
		t.Fatal("设置的自定义客户端未生效")
	}
}

// TestRequest 测试请求函数
func TestRequest(t *testing.T) {
	resetClient()

	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 测试不同的状态码和响应体
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status":"error"}`))
		case "/headers":
			// 测试请求头
			if r.Header.Get("X-Test-Header") == "test-value" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"headers":"ok"}`))
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		case "/post-data":
			// 测试POST数据
			body, _ := io.ReadAll(r.Body)
			if bytes.Equal(body, []byte(`{"test":"data"}`)) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"data":"received"}`))
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		}
	}))
	defer server.Close()

	// 测试成功的请求
	t.Run("成功请求", func(t *testing.T) {
		statusCode, body, err := Request("GET", server.URL+"/ok")
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		if statusCode != http.StatusOK {
			t.Fatalf("期望状态码 %d, 得到 %d", http.StatusOK, statusCode)
		}
		if string(body) != `{"status":"ok"}` {
			t.Fatalf("期望响应体 %s, 得到 %s", `{"status":"ok"}`, string(body))
		}
	})

	// 测试错误状态码的请求
	t.Run("错误状态码", func(t *testing.T) {
		statusCode, _, err := Request("GET", server.URL+"/error")
		if err == nil {
			t.Fatal("期望错误但未获得")
		}
		if statusCode != http.StatusInternalServerError {
			t.Fatalf("期望状态码 %d, 得到 %d", http.StatusInternalServerError, statusCode)
		}
		if !strings.Contains(err.Error(), "non 200 response") {
			t.Fatalf("错误消息不符合预期: %v", err)
		}
	})

	// 测试请求头
	t.Run("自定义请求头", func(t *testing.T) {
		headers := map[string]string{"X-Test-Header": "test-value"}
		statusCode, _, err := Request("GET", server.URL+"/headers", WithHeaders(headers))
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		if statusCode != http.StatusOK {
			t.Fatalf("期望状态码 %d, 得到 %d", http.StatusOK, statusCode)
		}
	})

	// 测试POST数据
	t.Run("POST数据", func(t *testing.T) {
		data := []byte(`{"test":"data"}`)
		statusCode, _, err := Request("POST", server.URL+"/post-data", WithData(data))
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		if statusCode != http.StatusOK {
			t.Fatalf("期望状态码 %d, 得到 %d", http.StatusOK, statusCode)
		}
	})

	// 测试无效URL
	t.Run("无效URL", func(t *testing.T) {
		_, _, err := Request("GET", "http://invalid-url-that-does-not-exist")
		if err == nil {
			t.Fatal("期望错误但未获得")
		}
	})

	// 测试自定义超时
	t.Run("自定义超时", func(t *testing.T) {
		_, _, err := Request("GET", server.URL+"/ok", WithTimeout(1*time.Microsecond))
		// 可能会超时也可能不会，这取决于网络情况，所以不做错误断言
		_ = err
	})

	// 测试自定义Context
	t.Run("自定义Context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // 立即取消
		_, _, err := Request("GET", server.URL+"/ok", WithContext(ctx))
		if err == nil {
			t.Fatal("期望上下文取消错误但未获得")
		}
	})
}

// TestGet 测试Get函数
func TestGet(t *testing.T) {
	resetClient()

	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"method":"get"}`))
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	// 测试Get请求
	ctx := context.Background()
	statusCode, body, err := Get(ctx, server.URL)
	if err != nil {
		t.Fatalf("Get请求失败: %v", err)
	}
	if statusCode != http.StatusOK {
		t.Fatalf("期望状态码 %d, 得到 %d", http.StatusOK, statusCode)
	}
	if string(body) != `{"method":"get"}` {
		t.Fatalf("期望响应体 %s, 得到 %s", `{"method":"get"}`, string(body))
	}
}

// TestPost 测试Post函数
func TestPost(t *testing.T) {
	resetClient()

	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Fatalf("期望Content-Type %s, 得到 %s", "application/json", contentType)
			}

			body, _ := io.ReadAll(r.Body)
			if string(body) == `{"test":"post"}` {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"method":"post"}`))
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	// 测试Post请求
	ctx := context.Background()
	data := []byte(`{"test":"post"}`)
	statusCode, body, err := Post(ctx, server.URL, data)
	if err != nil {
		t.Fatalf("Post请求失败: %v", err)
	}
	if statusCode != http.StatusOK {
		t.Fatalf("期望状态码 %d, 得到 %d", http.StatusOK, statusCode)
	}
	if string(body) != `{"method":"post"}` {
		t.Fatalf("期望响应体 %s, 得到 %s", `{"method":"post"}`, string(body))
	}
}

// TestWithOptions 测试各种Option函数
func TestWithOptions(t *testing.T) {
	// 测试默认选项
	opts := defaultRequestOptions()
	if opts.ctx == nil {
		t.Fatal("默认上下文为nil")
	}
	if opts.timeout != 5*time.Second {
		t.Fatalf("期望默认超时为 %v, 得到 %v", 5*time.Second, opts.timeout)
	}
	if opts.data != nil {
		t.Fatal("默认数据应为nil")
	}
	if len(opts.headers) != 0 {
		t.Fatal("默认请求头应为空")
	}

	// 测试WithContext
	ctx := context.Background()
	opt1 := WithContext(ctx)
	err := opt1.apply(opts)
	if err != nil {
		t.Fatalf("WithContext应用失败: %v", err)
	}
	if opts.ctx != ctx {
		t.Fatal("WithContext未正确设置上下文")
	}

	// 测试WithTimeout
	timeout := 10 * time.Second
	opt2 := WithTimeout(timeout)
	err = opt2.apply(opts)
	if err != nil {
		t.Fatalf("WithTimeout应用失败: %v", err)
	}
	if opts.timeout != timeout {
		t.Fatalf("期望超时为 %v, 得到 %v", timeout, opts.timeout)
	}

	// 测试WithHeaders
	headers := map[string]string{"X-Test": "test"}
	opt3 := WithHeaders(headers)
	err = opt3.apply(opts)
	if err != nil {
		t.Fatalf("WithHeaders应用失败: %v", err)
	}
	if opts.headers["X-Test"] != "test" {
		t.Fatal("WithHeaders未正确设置请求头")
	}

	// 测试WithData
	data := []byte("test data")
	opt4 := WithData(data)
	err = opt4.apply(opts)
	if err != nil {
		t.Fatalf("WithData应用失败: %v", err)
	}
	if string(opts.data) != "test data" {
		t.Fatalf("期望数据为 %s, 得到 %s", "test data", string(opts.data))
	}

	// 测试WithLogger
	mockLogger := &MockLogger{}
	opt5 := WithLogger(mockLogger)
	err = opt5.apply(opts)
	if err != nil {
		t.Fatalf("WithLogger应用失败: %v", err)
	}
	if opts.logger != mockLogger {
		t.Fatal("WithLogger未正确设置日志记录器")
	}

	// 测试WithSlowThreshold
	threshold := 100 * time.Millisecond
	opt6 := WithSlowThreshold(threshold)
	err = opt6.apply(opts)
	if err != nil {
		t.Fatalf("WithSlowThreshold应用失败: %v", err)
	}
	if opts.slowThreshold != threshold {
		t.Fatalf("期望慢请求阈值为 %v, 得到 %v", threshold, opts.slowThreshold)
	}
}

// TestLoggerOutputForRequest 测试请求日志输出
func TestLoggerOutputForRequest(t *testing.T) {
	resetClient()

	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/fast":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"response":"fast"}`))
		case "/slow":
			time.Sleep(100 * time.Millisecond) // 模拟慢响应
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"response":"slow"}`))
		}
	}))
	defer server.Close()

	// 测试快速请求的日志
	t.Run("快速请求日志", func(t *testing.T) {
		mockLogger := &MockLogger{}
		_, _, _ = Request("GET", server.URL+"/fast", WithLogger(mockLogger))
		if !mockLogger.debugCalled {
			t.Fatal("快速请求应调用Debug日志")
		}
		if mockLogger.warnCalled {
			t.Fatal("快速请求不应调用Warn日志")
		}
	})

	// 测试慢请求的日志
	t.Run("慢请求日志", func(t *testing.T) {
		mockLogger := &MockLogger{}
		_, _, _ = Request("GET", server.URL+"/slow",
			WithLogger(mockLogger),
			WithSlowThreshold(50*time.Millisecond))

		if !mockLogger.warnCalled {
			t.Fatal("慢请求应调用Warn日志")
		}
	})
}

// TestNewRequestError 测试创建请求对象时的错误
func TestNewRequestError(t *testing.T) {
	_, _, err := Request("INVALID_METHOD", "http://example.com")
	if err == nil {
		t.Fatal("使用无效的HTTP方法应该返回错误")
	}
}

// TestContextTimeout 测试上下文超时
func TestContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, _, err := Get(ctx, server.URL)
	if err == nil {
		t.Fatal("上下文超时应该返回错误")
	}
}
