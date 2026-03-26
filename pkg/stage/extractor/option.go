package extractor

import (
	"time"

	"github.com/KKKKKKKEM/grasp/pkg/extractors"
)

type stageOptions struct {
	fallback        extractors.Opts // proxy/timeout/retry/headers 等兜底值
	inputKey        string          // 从 rc.Values 中读取 Task 的 key，默认为 "task"
	nextStageName   string
	maxRounds       int
	defaultSelector extractors.Selector
}

type Option func(*stageOptions)

func WithInputKey(inputKey string) Option {
	return func(o *stageOptions) { o.inputKey = inputKey }
}

func WithProxy(proxyURL string) Option {
	return func(o *stageOptions) { o.fallback.Proxy = proxyURL }
}

func WithEnvProxy() Option {
	return func(o *stageOptions) { o.fallback.Proxy = "env" }
}

func WithNextStage(stageName string) Option {
	return func(o *stageOptions) { o.nextStageName = stageName }
}

func WithRetry(maxAttempts int, interval time.Duration) Option {
	return func(o *stageOptions) { o.fallback.Retry = maxAttempts }
}

func WithTimeout(d time.Duration) Option {
	return func(o *stageOptions) { o.fallback.Timeout = d }
}

func WithHeaders(headers map[string]string) Option {
	return func(o *stageOptions) { o.fallback.Headers = headers }
}

func WithHeader(key, value string) Option {
	return func(o *stageOptions) {
		if o.fallback.Headers == nil {
			o.fallback.Headers = make(map[string]string)
		}
		o.fallback.Headers[key] = value
	}
}
