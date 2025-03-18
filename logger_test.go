package httptool

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"
)

// TestLoggerLevels 测试不同日志级别的输出
func TestLoggerLevels(t *testing.T) {
	// 创建一个临时文件用于捕获日志输出
	tmpfile, err := os.CreateTemp("", "logger_test_*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// 创建测试用的logger
	testLogger := New(log.New(tmpfile, "", 0), Config{
		LogLevel: Debug,
		Colorful: false,
	})

	ctx := context.Background()

	// 测试Debug级别
	testLogger.Debug(ctx, "debug message", "key", "value")
	testLogger.Info(ctx, "info message", "key", "value")
	testLogger.Warn(ctx, "warn message", "key", "value")
	testLogger.Error(ctx, "error message", "key", "value")

	// 读取日志文件内容
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// 验证日志内容
	logContent := string(content)
	expectedMessages := []string{
		"[debug] debug message",
		"[info] info message",
		"[warn] warn message",
		"[error] error message",
	}

	for _, msg := range expectedMessages {
		if !strings.Contains(logContent, msg) {
			t.Errorf("日志中缺少预期消息: %s", msg)
		}
	}
}

// TestLoggerColorful 测试彩色输出
func TestLoggerColorful(t *testing.T) {
	// 创建一个临时文件用于捕获日志输出
	tmpfile, err := os.CreateTemp("", "logger_test_*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// 创建测试用的logger，启用彩色输出
	testLogger := New(log.New(tmpfile, "", 0), Config{
		LogLevel: Debug,
		Colorful: true,
	})

	ctx := context.Background()
	testLogger.Debug(ctx, "colored debug message")

	// 读取日志文件内容
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// 验证是否包含颜色代码
	logContent := string(content)
	if !strings.Contains(logContent, Green) || !strings.Contains(logContent, Reset) {
		t.Error("彩色输出未正确启用")
	}
}

// TestLoggerLevelFilter 测试日志级别过滤
func TestLoggerLevelFilter(t *testing.T) {
	// 创建一个临时文件用于捕获日志输出
	tmpfile, err := os.CreateTemp("", "logger_test_*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// 创建测试用的logger，设置日志级别为Warn
	testLogger := New(log.New(tmpfile, "", 0), Config{
		LogLevel: Warn,
		Colorful: false,
	})

	ctx := context.Background()

	// 记录不同级别的日志
	testLogger.Debug(ctx, "debug message")
	testLogger.Info(ctx, "info message")
	testLogger.Warn(ctx, "warn message")
	testLogger.Error(ctx, "error message")

	// 读取日志文件内容
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// 验证日志内容
	logContent := string(content)

	// Debug和Info消息不应该出现
	if strings.Contains(logContent, "[debug]") {
		t.Error("Debug消息不应该被记录")
	}
	if strings.Contains(logContent, "[info]") {
		t.Error("Info消息不应该被记录")
	}

	// Warn和Error消息应该出现
	if !strings.Contains(logContent, "[warn]") {
		t.Error("Warn消息应该被记录")
	}
	if !strings.Contains(logContent, "[error]") {
		t.Error("Error消息应该被记录")
	}
}

// TestLoggerCallerInfo 测试调用者信息
func TestLoggerCallerInfo(t *testing.T) {
	// 创建一个临时文件用于捕获日志输出
	tmpfile, err := os.CreateTemp("", "logger_test_*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// 创建测试用的logger
	testLogger := New(log.New(tmpfile, "", 0), Config{
		LogLevel: Debug,
		Colorful: false,
	})

	ctx := context.Background()
	testLogger.Debug(ctx, "test message")

	// 读取日志文件内容
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// 验证调用者信息
	logContent := string(content)
	if !strings.Contains(logContent, "logger_test.go") {
		t.Error("日志中缺少调用者文件名")
	}
}

// TestLoggerMode 测试日志模式切换
func TestLoggerMode(t *testing.T) {
	// 创建一个临时文件用于捕获日志输出
	tmpfile, err := os.CreateTemp("", "logger_test_*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// 创建测试用的logger
	testLogger := New(log.New(tmpfile, "", 0), Config{
		LogLevel: Debug,
		Colorful: false,
	})

	ctx := context.Background()

	// 切换到Silent模式
	testLogger = testLogger.LogMode(Silent)
	testLogger.Debug(ctx, "debug message")
	testLogger.Info(ctx, "info message")
	testLogger.Warn(ctx, "warn message")
	testLogger.Error(ctx, "error message")

	// 读取日志文件内容
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// 验证Silent模式下没有日志输出
	if len(content) > 0 {
		t.Error("Silent模式下不应该有日志输出")
	}
}
