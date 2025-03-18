package httptool

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	client *http.Client
	once   sync.Once
)

// GetHttpClient 获取全局HTTP客户端
func GetHttpClient() *http.Client {
	if client != nil {
		// 因为需要支持自定义client, 所以虽然用了once.Do但是还是先判断一下client有没有实例化
		return client
	}
	once.Do(func() {
		tr := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			MaxIdleConnsPerHost:   50,
			MaxConnsPerHost:       50,
			ForceAttemptHTTP2:     true,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
		client = &http.Client{Transport: tr}
	})
	return client
}

// SetHttpClient 提供传入自定义HttpClient方法
func SetHttpClient(c *http.Client) {
	client = c
}

func Request(method string, url string, options ...Option) (httpStatusCode int, respBody []byte, err error) {
	start := time.Now()
	reqOpts := defaultRequestOptions() // 默认的请求选项
	for _, opt := range options {      // 在reqOpts上应用通过options设置的选项
		err = opt.apply(reqOpts)
		if err != nil {
			return
		}
	}

	// 创建请求对象
	req, err := http.NewRequest(method, url, bytes.NewReader(reqOpts.data))
	if err != nil {
		return
	}
	reqOpts.ctx, _ = context.WithTimeout(reqOpts.ctx, reqOpts.timeout) // 给 Request 设置Timeout
	req = req.WithContext(reqOpts.ctx)
	defer req.Body.Close()

	if len(reqOpts.headers) != 0 { // 设置请求头
		for key, value := range reqOpts.headers {
			req.Header.Add(key, value)
		}
	}
	// 发起请求
	client := GetHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	// 记录请求日志
	dur := time.Since(start)
	defer func() {
		if reqOpts.slowThreshold > 0 && dur >= reqOpts.slowThreshold { // 超过 阈值 返回, 记一条 Warn 日志
			reqOpts.logger.Warn(reqOpts.ctx, "HTTP_REQUEST_SLOW_LOG", "method", method, "url", url, "body", reqOpts.data, "reply", respBody, "err", err, "dur/ms", dur)
		} else {
			reqOpts.logger.Debug(reqOpts.ctx, "HTTP_REQUEST_DEBUG_LOG", "method", method, "url", url, "body", string(reqOpts.data), "reply", string(respBody), "err", err, "dur/ms", dur)
		}
	}()

	httpStatusCode = resp.StatusCode
	if httpStatusCode != http.StatusOK {
		// 返回非 200 时Go的 http 库不回返回error, 这里处理成error 调用方好判断
		err = errors.New(fmt.Sprintf("non 200 response, response code: %d", httpStatusCode))
		return
	}

	respBody, _ = io.ReadAll(resp.Body)
	return
}

// Get 发起GET请求
func Get(ctx context.Context, url string, options ...Option) (httpStatusCode int, respBody []byte, err error) {
	options = append(options, WithContext(ctx))
	return Request("GET", url, options...)
}

// Post 发起POST请求
func Post(ctx context.Context, url string, data []byte, options ...Option) (httpStatusCode int, respBody []byte, err error) {
	// 默认自带Header Content-Type: application/json 可通过 传递 WithHeaders 增加或者覆盖Header信息
	defaultHeader := map[string]string{"Content-Type": "application/json"}
	var newOptions []Option
	newOptions = append(newOptions, WithHeaders(defaultHeader), WithData(data), WithContext(ctx))
	newOptions = append(newOptions, options...)

	httpStatusCode, respBody, err = Request("POST", url, newOptions...)
	return
}

// 针对可选的HTTP请求配置项，模仿gRPC使用的Options设计模式实现
type requestOption struct {
	ctx           context.Context
	timeout       time.Duration
	data          []byte
	headers       map[string]string
	logger        Interface
	slowThreshold time.Duration // 慢请求阈值
}

type Option interface {
	apply(option *requestOption) error
}

type optionFunc func(option *requestOption) error

func (f optionFunc) apply(opts *requestOption) error {
	return f(opts)
}

func defaultRequestOptions() *requestOption {
	return &requestOption{ // 默认请求选项
		ctx:     context.Background(),
		timeout: 5 * time.Second,
		data:    nil,
		headers: map[string]string{},
		logger:  Default,
	}
}

func WithContext(ctx context.Context) Option {
	return optionFunc(func(opts *requestOption) (err error) {
		opts.ctx = ctx
		return
	})
}

func WithTimeout(timeout time.Duration) Option {
	return optionFunc(func(opts *requestOption) (err error) {
		opts.timeout, err = timeout, nil
		return
	})
}

func WithHeaders(headers map[string]string) Option {
	return optionFunc(func(opts *requestOption) (err error) {
		for k, v := range headers {
			opts.headers[k] = v
		}
		return
	})
}

func WithData(data []byte) Option {
	return optionFunc(func(opts *requestOption) (err error) {
		opts.data, err = data, nil
		return
	})
}

func WithLogger(l Interface) Option {
	return optionFunc(func(opts *requestOption) (err error) {
		opts.logger, err = l, nil
		return
	})
}

// WithSlowThreshold 设置慢请求阈值 单位:毫秒
func WithSlowThreshold(threshold time.Duration) Option {
	return optionFunc(func(opts *requestOption) (err error) {
		opts.slowThreshold, err = threshold, nil
		return
	})
}
