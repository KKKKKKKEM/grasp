package download

import (
	"fmt"
	"sync"

	"github.com/KKKKKKKEM/flowkit/core"
)

type batchResult struct {
	task   *Task
	result *Result
	err    error
}

type DirectDownloadStage struct {
	stageName string
	opts      stageOptions
}

func (s *DirectDownloadStage) Name() string {
	return s.stageName
}

func applyFallback(task *Task, fb *Opts, headers map[string]string) {
	if task.Opts == nil {
		task.Opts = &Opts{}
	}
	if task.Opts.Proxy == "" {
		task.Opts.Proxy = fb.Proxy
	}
	if task.Opts.Timeout == 0 {
		task.Opts.Timeout = fb.Timeout
	}
	if task.Opts.Retry == 0 {
		task.Opts.Retry = fb.Retry
	}
	if task.Opts.RetryInterval == 0 {
		task.Opts.RetryInterval = fb.RetryInterval
	}
	if headers != nil {
		if task.Request.Header == nil {
			task.Request.Header = make(map[string][]string)
		}
		for k, v := range headers {
			if _, exists := task.Request.Header[k]; !exists {
				task.Request.Header[k] = []string{v}
			}
		}
	}
}

func (s *DirectDownloadStage) Run(rc *core.Context) core.StageResult {
	tasks, err := s.loadTasks(rc)
	if err != nil {
		return core.StageResult{Status: core.StageFailed, Err: err}
	}

	o := s.opts
	for _, task := range tasks {
		applyFallback(task, &o.fallback, o.headers)
	}

	if len(tasks) == 1 {
		return s.downloadOne(rc, tasks[0])
	}
	return s.downloadBatch(rc, tasks)
}

func (s *DirectDownloadStage) loadTasks(rc *core.Context) ([]*Task, error) {
	inputKey := s.opts.inputKey
	if inputKey == "" {
		inputKey = "tasks"
	}

	val, ok := rc.State.Get(inputKey)
	if !ok {
		return nil, fmt.Errorf("task not found in rc.State[%q]", inputKey)
	}

	switch v := val.(type) {
	case []*Task:
		return v, nil
	case *Task:
		return []*Task{v}, nil
	default:
		return nil, fmt.Errorf("rc.State[%q] must be *Task or []*Task, got %T", inputKey, val)
	}
}

func (s *DirectDownloadStage) downloadOne(rc *core.Context, task *Task) core.StageResult {
	dl := NewHTTPDownloader()
	result, err := dl.Download(rc, task)
	if err != nil {
		return core.StageResult{Status: core.StageFailed, Err: err}
	}
	if task.OnComplete != nil {
		task.OnComplete(result)
	}
	return core.StageResult{
		Status:  core.StageSuccess,
		Next:    s.opts.nextStageName,
		Outputs: map[string]any{"download_results": []*Result{result}},
	}
}

func (s *DirectDownloadStage) downloadBatch(rc *core.Context, tasks []*Task) core.StageResult {
	resultsCh := make(chan batchResult, len(tasks))
	var wg sync.WaitGroup

	for _, task := range tasks {
		wg.Add(1)
		go func(t *Task) {
			defer wg.Done()
			dl := NewHTTPDownloader()
			res, err := dl.Download(rc, t)
			if err == nil && t.OnComplete != nil {
				t.OnComplete(res)
			}
			resultsCh <- batchResult{task: t, result: res, err: err}
		}(task)
	}

	wg.Wait()
	close(resultsCh)

	var results []*Result
	for br := range resultsCh {
		if br.err != nil {
			return core.StageResult{
				Status: core.StageFailed,
				Err:    fmt.Errorf("download failed for %s: %w", br.task.Request.URL, br.err),
			}
		}
		results = append(results, br.result)
	}

	return core.StageResult{
		Status:  core.StageSuccess,
		Next:    s.opts.nextStageName,
		Outputs: map[string]any{"download_results": results},
	}
}

func NewStage(name string, options ...Option) *DirectDownloadStage {
	s := &DirectDownloadStage{
		stageName: name,
	}
	for _, opt := range options {
		opt(&s.opts)
	}
	return s
}
