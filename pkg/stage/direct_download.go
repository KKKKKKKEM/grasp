package stage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/KKKKKKKEM/grasp/pkg/core"
	"github.com/KKKKKKKEM/grasp/pkg/downloader"
	"github.com/KKKKKKKEM/grasp/pkg/downloader/http"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

var (
	sharedProgressOnce sync.Once
	sharedProgress     *mpb.Progress
)

func getSharedProgress() *mpb.Progress {
	sharedProgressOnce.Do(func() {
		sharedProgress = mpb.New(
			mpb.WithRefreshRate(120 * time.Millisecond),
		)
	})
	return sharedProgress
}

type stageOptions struct {
	progressBar   bool
	proxy         string
	timeout       time.Duration
	retry         int
	retryInterval time.Duration
}

type Option func(*stageOptions)

func WithProgressBar() Option {
	return func(o *stageOptions) { o.progressBar = true }
}

func WithProxy(proxyURL string) Option {
	return func(o *stageOptions) { o.proxy = proxyURL }
}

func WithEnvProxy() Option {
	return func(o *stageOptions) { o.proxy = "env" }
}

func WithTimeout(d time.Duration) Option {
	return func(o *stageOptions) { o.timeout = d }
}

func WithRetry(maxAttempts int, interval time.Duration) Option {
	return func(o *stageOptions) {
		o.retry = maxAttempts
		o.retryInterval = interval
	}
}

type DirectDownloadStage struct {
	Task *downloader.Task
	opts stageOptions
}

func (s *DirectDownloadStage) Do(ctx context.Context) (core.Stage, error) {
	task := s.Task
	o := s.opts
	if task == nil {
		return nil, fmt.Errorf("task is nil")
	}
	if task.Opts == nil {
		task.Opts = &downloader.Opts{}
	}

	if task.Proxy == "" && o.proxy != "" {
		task.Proxy = o.proxy
	}
	if task.Timeout == 0 && o.timeout > 0 {
		task.Timeout = o.timeout
	}
	if task.Retry == 0 && o.retry > 0 {
		task.Retry = o.retry
	}
	if task.RetryInterval == 0 && o.retryInterval > 0 {
		task.RetryInterval = o.retryInterval
	}

	if o.progressBar {
		p := getSharedProgress()
		savePath, err := task.GetSavePath()
		if err != nil {
			return nil, fmt.Errorf("failed to get save path: %w", err)
		}

		bar := p.AddBar(0,
			mpb.PrependDecorators(
				decor.Name(savePath+" ", decor.WCSyncWidth),
				decor.Counters(decor.SizeB1024(0), "% .2f / % .2f"),
			),
			mpb.AppendDecorators(
				decor.OnComplete(
					decor.EwmaETA(decor.ET_STYLE_GO, 30, decor.WCSyncWidth),
					"done",
				),
				decor.Name(" "),
				decor.EwmaSpeed(decor.SizeB1024(0), "% .2f", 30, decor.WCSyncWidth),
			),
		)

		var (
			mu         sync.Mutex
			knownTotal int64
			lastBytes  int64
		)

		origProgress := task.OnProgress
		task.OnProgress = func(downloaded, total int64) {
			mu.Lock()
			defer mu.Unlock()

			if total > 0 && total != knownTotal {
				knownTotal = total
				bar.SetTotal(total, false)
			}

			delta := downloaded - lastBytes
			if delta > 0 {
				bar.EwmaIncrInt64(delta, 120*time.Millisecond)
				lastBytes = downloaded
			}

			if origProgress != nil {
				origProgress(downloaded, total)
			}
		}

		origComplete := task.OnComplete
		task.OnComplete = func(result *downloader.DownloadResult) {
			mu.Lock()
			bar.SetTotal(-1, true)
			mu.Unlock()

			if origComplete != nil {
				origComplete(result)
			}
		}

		dl := http.NewSimpleHTTPDownloader()
		result, err := dl.Download(ctx, task)
		if err != nil {
			bar.Abort(true)
			p.Wait()
			return nil, err
		}
		if task.OnComplete != nil {
			task.OnComplete(result)
		}

		p.Wait()
		return nil, nil
	}

	dl := http.NewSimpleHTTPDownloader()
	_, err := dl.Download(ctx, task)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func NewDirectDownloadStage(task *downloader.Task, options ...Option) *DirectDownloadStage {
	s := &DirectDownloadStage{Task: task}
	for _, opt := range options {
		opt(&s.opts)
	}
	return s
}
