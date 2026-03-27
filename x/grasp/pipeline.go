package grasp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/KKKKKKKEM/flowkit/builtin/download"
	"github.com/KKKKKKKEM/flowkit/builtin/extract"
	"github.com/KKKKKKKEM/flowkit/core"
	"github.com/google/uuid"
)

type GraspReport struct {
	Success     bool
	DurationMs  int64
	Rounds      int
	ParsedItems int
	Downloaded  []*download.Result
}

type Pipeline struct {
	extractor        *extract.Stage
	downloader       *download.DirectDownloadStage
	defaultSelector  SelectFunc
	defaultTransform TransformFunc
	reporter         ProgressReporter
}

func NewGraspPipeline(opts ...Option) *Pipeline {
	p := &Pipeline{}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Pipeline) Run(ctx context.Context, task *Task) (*GraspReport, error) {
	start := time.Now()
	report := &GraspReport{}

	allDirect, rounds, err := p.runExtract(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("extract: %w", err)
	}
	report.Rounds = rounds
	report.ParsedItems = len(allDirect)

	selected, err := task.resolveSelector(p.defaultSelector)(ctx, allDirect)
	if err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	dlTasks, err := p.buildDownloadTasks(ctx, selected, task)
	if err != nil {
		return nil, fmt.Errorf("transform: %w", err)
	}

	results, err := p.runDownload(ctx, dlTasks)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}

	report.Downloaded = results
	report.Success = true
	report.DurationMs = time.Since(start).Milliseconds()
	return report, nil
}

func (p *Pipeline) runExtract(ctx context.Context, task *Task) ([]extract.ParseItem, int, error) {
	maxRounds := task.Extract.MaxRounds
	if maxRounds <= 0 {
		maxRounds = 1
	}
	concurrency := task.Extract.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	extractOpts := task.toExtractOpts()
	var allDirect []extract.ParseItem
	queue := make([]string, len(task.URLs))
	copy(queue, task.URLs)

	for round := 0; round < maxRounds && len(queue) > 0; round++ {
		items, next, err := p.extractRound(ctx, queue, task.Extract.ForcedParser, extractOpts, concurrency)
		if err != nil {
			return nil, round + 1, err
		}
		allDirect = append(allDirect, items...)
		queue = next
		if round == 0 {
			task.Extract.ForcedParser = ""
		}
	}

	return allDirect, maxRounds, nil
}

func (p *Pipeline) extractRound(
	ctx context.Context,
	urls []string,
	forcedParser string,
	opts *extract.Opts,
	concurrency int,
) (direct []extract.ParseItem, nextQueue []string, err error) {
	type result struct {
		items []extract.ParseItem
		err   error
	}

	sem := make(chan struct{}, concurrency)
	results := make([]result, len(urls))
	var wg sync.WaitGroup

	for i, url := range urls {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, u string) {
			defer wg.Done()
			defer func() { <-sem }()

			rc := core.NewRunContext(ctx, uuid.NewString())
			rc.WithValue("task", &extract.Task{
				URL:          u,
				Opts:         opts,
				ForcedParser: forcedParser,
			})
			sr := p.extractor.Run(rc)
			if sr.IsFailed() {
				results[idx] = result{err: sr.Err}
				return
			}
			items, _ := sr.Outputs["items"].([]extract.ParseItem)
			results[idx] = result{items: items}
		}(i, url)
	}
	wg.Wait()

	for _, r := range results {
		if r.err != nil {
			return nil, nil, r.err
		}
		for _, item := range r.items {
			if item.IsDirect {
				direct = append(direct, item)
			} else {
				nextQueue = append(nextQueue, item.URI)
			}
		}
	}
	return direct, nextQueue, nil
}

func (p *Pipeline) buildDownloadTasks(ctx context.Context, items []extract.ParseItem, task *Task) ([]*download.Task, error) {
	transformFn := task.resolveTransform(p.defaultTransform)
	baseOpts := task.toDownloadOpts()

	tasks := make([]*download.Task, 0, len(items))
	for _, item := range items {
		t, err := transformFn(ctx, item, baseOpts)
		if err != nil {
			return nil, fmt.Errorf("transform %q: %w", item.URI, err)
		}
		if p.reporter != nil {
			p.reporter.Track(t)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (p *Pipeline) runDownload(ctx context.Context, tasks []*download.Task) ([]*download.Result, error) {
	rc := core.NewRunContext(ctx, uuid.NewString())
	rc.WithValue("tasks", tasks)

	sr := p.downloader.Run(rc)
	if sr.IsFailed() {
		return nil, sr.Err
	}

	results, _ := sr.Outputs["download_results"].([]*download.Result)
	return results, nil
}
