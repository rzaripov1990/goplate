package client

import (
	"context"
	"encoding/json"
	"goplate/pkg/trace_logger"
	"log/slog"
	"time"

	trace_context "github.com/rzaripov1990/trace_ctx"

	"github.com/valyala/fasthttp"
)

type (
	Http struct {
		withRetry  bool
		retryCount int
		client     *fasthttp.Client
		baseUrl    string
	}

	Config struct {
		BaseUrl             string
		ReadTimeout         time.Duration
		WriteTimeout        time.Duration
		MaxIdleConns        int
		MaxIdleConnsPerHost int
		RetryIfCount        int
	}

	Request struct {
		client  *fasthttp.Client
		headers map[string]string
		uri     string
	}
)

func New(cfg Config) *Http {
	fhc := &Http{
		withRetry:  cfg.RetryIfCount >= 1,
		baseUrl:    cfg.BaseUrl,
		retryCount: cfg.RetryIfCount,
	}

	fhc.client = &fasthttp.Client{
		// Name:                     cfg.App.Name + " " + cfg.App.Version,
		// NoDefaultUserAgentHeader: true,
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      4096,
			DNSCacheDuration: time.Hour,
		}).Dial,
		DialDualStack:   false,
		MaxConnsPerHost: cfg.MaxIdleConnsPerHost,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		RetryIf:         fhc.retry,
	}

	return fhc
}

func (fhc *Http) NewRequest(path string, headers map[string]string) *Request {
	value := &Request{
		headers: headers,
		uri:     fhc.baseUrl + path,
		client:  fhc.client,
	}
	return value
}

func Get[T_response any](ctx context.Context, log *slog.Logger, req *Request) (resp *T_response, err error) {
	ctx = trace_context.WithTraceID(ctx)

	fastreq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(fastreq)

	fastresp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(fastresp)

	fastreq.SetRequestURI(req.uri)

	if req.headers == nil {
		fastreq.Header.Set(fasthttp.HeaderContentType, "application/json")
	} else {
		for key := range req.headers {
			fastreq.Header.Set(key, req.headers[key])
		}
	}

	var slogValues []slog.Attr
	slogValues, err = req.do(fastreq, fastresp, &resp)

	trace_logger.L(ctx, log).LogAttrs(ctx, slog.LevelDebug, "logging request", slogValues...)
	return
}

func (fr *Request) do(fastreq *fasthttp.Request, fastresp *fasthttp.Response, resp any) (slogValues []slog.Attr, err error) {
	var (
		start = time.Now()
	)

	slogValues = append(slogValues,
		slog.String("request.url", string(fastreq.RequestURI())),
		slog.String("request.method", string(fastreq.Header.Method())),
	)
	for _, value := range fastreq.Header.PeekKeys() {
		slogValues = append(slogValues, slog.String("request.header."+string(value), string(fastreq.Header.Peek(string(value)))))
	}

	err = fr.client.Do(fastreq, fastresp)
	duration := time.Since(start)
	if err != nil {
		slogValues = append(slogValues,
			slog.String("response.error", err.Error()),
			slog.Duration("response.duration_nanosec", duration),
			slog.String("response.duration", duration.String()),
		)
		return
	}

	slogValues = append(slogValues,
		slog.Duration("response.duration_nanosec", duration),
		slog.String("response.duration", duration.String()),
		slog.String("response.content_type", string(fastresp.Header.Peek(fasthttp.HeaderContentType))),
		slog.Int("response.code", fastresp.StatusCode()),
	)

	body := fastresp.Body()
	if len(body) > 0 {
		if err = json.Unmarshal(body, resp); err == nil {
			slogValues = append(slogValues, slog.Any("response.body", resp))
		} else {
			slogValues = append(slogValues, slog.String("response.error", err.Error()))
			return
		}
	}

	return
}

func (fhc *Http) retry(_ *fasthttp.Request) bool {
	if fhc.withRetry {
		if fhc.retryCount < 0 {
			fhc.withRetry = false
		}
		fhc.retryCount--
	}

	return fhc.withRetry
}
