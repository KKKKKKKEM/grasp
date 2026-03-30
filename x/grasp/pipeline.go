package grasp

import (
	"fmt"
	"sync"
	"time"

	"github.com/KKKKKKKEM/flowkit"
	"github.com/KKKKKKKEM/flowkit/core"
	"github.com/KKKKKKKEM/flowkit/pipeline"
	"github.com/KKKKKKKEM/flowkit/x/download"
	"github.com/KKKKKKKEM/flowkit/x/extract"
	"github.com/KKKKKKKEM/flowkit/x/grasp/sites/pexels"
	"github.com/google/uuid"
)

type Report struct {
	Success     bool               `json:"success"`
	DurationMs  int64              `json:"duration_ms"`
	Rounds      int                `json:"rounds"`
	ParsedItems int                `json:"parsed_items"`
	Downloaded  []*download.Result `json:"downloaded"`
}

type Pipeline struct {
	flowkit.App[*Task, *Report]
	*pipeline.LinearPipeline
	extractor         *extract.Stage
	downloader        *download.Stage
	defaultSelector   SelectFunc
	defaultTransform  TransformFunc
	interactionPlugin core.InteractionPlugin
	trackerProvider   core.TrackerProvider
}

var _ core.Pipeline = (*Pipeline)(nil)

func NewGraspPipeline(opts ...Option) *Pipeline {
	extractor := extract.NewStage("extractor")
	extractor.Mount(&pexels.APIParser{})

	downloader := download.NewStage("download")
	p := &Pipeline{
		LinearPipeline:    pipeline.NewLinearPipeline(),
		extractor:         extractor,
		downloader:        downloader,
		trackerProvider:   NewMPBTrackerProvider(),
		interactionPlugin: &CLIInteractionPlugin{},
	}
	for _, opt := range opts {
		opt(p)
	}
	p.App = flowkit.NewApp(p.Invoke)
	return p
}

func (p *Pipeline) CLI(opts ...flowkit.CLIOption[*Task, *Report]) error {
	return p.App.CLI(append([]flowkit.CLIOption[*Task, *Report]{
		flowkit.WithCLIBuilder[*Task, *Report](buildCLI),
		flowkit.WithTrackerProvider[*Task, *Report](p.trackerProvider),
		flowkit.WithInteractionPlugin[*Task, *Report](p.interactionPlugin),
	}, opts...)...)
}

func (p *Pipeline) Serve(addr string, opts ...flowkit.ServeOption[*Task, *Report]) error {
	return p.App.Serve(addr, opts...)
}

func (p *Pipeline) Run(rc *core.Context, _ string) (*core.Report, error) {
	report := &core.Report{
		Mode:         core.ModeLinear,
		TraceID:      rc.Runtime.TraceID,
		StageOrder:   []string{"grasp"},
		StageResults: make(map[string]core.StageResult),
	}
	start := time.Now()

	v, ok := rc.State.Get("task")
	task, ok := v.(*Task)
	if !ok {
		return fail(report, start, fmt.Errorf("rc.State[\"task\"] missing or wrong type"))
	}

	graspReport, err := p.run(rc, task)
	if err != nil {
		return fail(report, start, err)
	}

	report.StageResults["grasp"] = core.StageResult{
		Status:  core.StageSuccess,
		Outputs: map[string]any{"report": graspReport},
	}
	report.Success = true
	report.DurationMs = time.Since(start).Milliseconds()
	return report, nil
}

func (p *Pipeline) Invoke(rc *core.Context, task *Task) (*Report, error) {
	rc.State.Set("task", task)
	runReport, err := p.Run(rc, "grasp")
	if err != nil {
		return nil, err
	}
	result, ok := runReport.StageResults["grasp"].Outputs["report"].(*Report)
	if !ok {
		return nil, fmt.Errorf("unexpected report type in StageResults")
	}
	return result, nil
}

func fail(report *core.Report, start time.Time, err error) (*core.Report, error) {
	report.StageResults["grasp"] = core.StageResult{
		Status: core.StageFailed,
		Err:    err,
	}
	report.Success = false
	report.DurationMs = time.Since(start).Milliseconds()
	return report, err
}

func (p *Pipeline) run(rc *core.Context, task *Task) (*Report, error) {
	start := time.Now()
	report := &Report{}

	allDirect, rounds, err := p.runExtract(rc, task)
	if err != nil {
		return nil, fmt.Errorf("extract: %w", err)
	}
	report.Rounds = rounds
	report.ParsedItems = len(allDirect)

	selected, err := p.selectItems(rc, task, allDirect)
	if err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	dlTasks, err := p.buildDownloadTasks(rc, selected, task)
	if err != nil {
		return nil, fmt.Errorf("transform: %w", err)
	}

	results, err := p.runDownload(rc, dlTasks)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}

	report.Downloaded = results
	report.Success = true
	report.DurationMs = time.Since(start).Milliseconds()
	return report, nil
}

func (p *Pipeline) selectItems(rc *core.Context, task *Task, items []extract.ParseItem) ([]extract.ParseItem, error) {
	if task.Selector != nil {
		return task.Selector(rc, items)
	}

	i := core.Interaction{Type: core.InteractionTypeSelect, Payload: items, Message: "Please select items to download"}
	interactionPlugin := rc.Runtime.InteractionPlugin
	if interactionPlugin == nil {
		interactionPlugin = p.interactionPlugin
	}

	var indices []int
	if interactionPlugin != nil {
		result, err := interactionPlugin.Interact(rc, i)
		if err != nil {
			return nil, err
		}

		result, err = interactionPlugin.FormatResult(rc, i, result)
		if err != nil {
			return nil, err
		}
		indices, err = toIntSlice(result.Answer)
		if err != nil {
			return nil, fmt.Errorf("select interaction: invalid answer: %w", err)
		}
	} else {
		return task.resolveSelector(p.defaultSelector)(rc, items)
	}

	if len(indices) == 0 {
		return nil, fmt.Errorf("no items selected")
	}

	var selected []extract.ParseItem

	for _, index := range indices {
		selected = append(selected, items[index])
	}

	return selected, nil

}

func toIntSlice(v any) ([]int, error) {
	switch val := v.(type) {
	case []int:
		return val, nil
	case []any:
		out := make([]int, 0, len(val))
		for _, item := range val {
			f, ok := item.(float64)
			if !ok {
				return nil, fmt.Errorf("expected number, got %T", item)
			}
			out = append(out, int(f))
		}
		return out, nil
	default:
		return nil, fmt.Errorf("expected []int or []any, got %T", v)
	}
}

func (p *Pipeline) runExtract(rc *core.Context, task *Task) ([]extract.ParseItem, int, error) {
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
		items, next, err := p.extractRound(rc, queue, task.Extract.ForcedParser, extractOpts, concurrency)
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
	rc *core.Context,
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

			child := rc.Fork(uuid.NewString())
			typed, err := p.extractor.Exec(child, &extract.Task{
				URL:          u,
				Opts:         opts,
				ForcedParser: forcedParser,
			})
			if err != nil {
				results[idx] = result{err: err}
				return
			}

			results[idx] = result{items: typed.Output}

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

func (p *Pipeline) buildDownloadTasks(rc *core.Context, items []extract.ParseItem, task *Task) ([]*download.Task, error) {
	transformFn := task.resolveTransform(p.defaultTransform)
	baseOpts := task.toDownloadOpts()

	trackerBuilder := rc.Runtime.TrackerProvider
	if trackerBuilder == nil {
		trackerBuilder = p.trackerProvider
	}

	tasks := make([]*download.Task, 0, len(items))
	for _, item := range items {
		t, err := transformFn(rc, item, baseOpts)
		if err != nil {
			return nil, fmt.Errorf("transform %q: %w", item.URI, err)
		}
		if trackerBuilder != nil {
			key := item.URI
			if item.Name != "" {
				key = item.Name
			}
			tracker := trackerBuilder.Track(key, map[string]any{"total": 0})
			bridgeDownloadTask(t, tracker)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (p *Pipeline) runDownload(rc *core.Context, tasks []*download.Task) ([]*download.Result, error) {
	child := rc.Fork(uuid.NewString())
	typed, err := p.downloader.Exec(child, tasks)
	if err != nil {
		return nil, err
	}
	return typed.Output, nil
}

func bridgeDownloadTask(task *download.Task, tracker core.Tracker) {
	origProgress := task.OnProgress
	task.OnProgress = func(downloaded, total int64) {
		tracker.Update(map[string]any{"current": downloaded, "total": total})
		tracker.Flush()
		if origProgress != nil {
			origProgress(downloaded, total)
		}
	}

	origComplete := task.OnComplete
	task.OnComplete = func(result *download.Result) {
		tracker.Done()
		if origComplete != nil {
			origComplete(result)
		}
	}
}
