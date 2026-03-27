package pipeline

import (
	"fmt"
	"time"

	"github.com/KKKKKKKEM/flowkit/core"
)

// LinearPipeline 是顺序执行的线性管道
type LinearPipeline struct {
	stages map[string]core.Stage
	mw     []core.Middleware
}

func NewLinearPipeline() *LinearPipeline {
	return &LinearPipeline{
		stages: make(map[string]core.Stage),
	}
}

func (lp *LinearPipeline) Mode() core.Mode {
	return core.ModeLinear
}

func (lp *LinearPipeline) Register(stages ...core.Stage) {
	for _, s := range stages {
		lp.stages[s.Name()] = s
	}
}

func (lp *LinearPipeline) Use(mw ...core.Middleware) *LinearPipeline {
	lp.mw = append(lp.mw, mw...)
	return lp
}

func (lp *LinearPipeline) Run(rc *core.Context, entry string) (*core.Report, error) {
	report := &core.Report{
		Mode:         core.ModeLinear,
		TraceID:      rc.TraceID,
		StageOrder:   []string{},
		StageResults: make(map[string]core.StageResult),
		DurationMs:   0,
	}

	start := time.Now()

	st, ok := lp.stages[entry]
	if !ok {
		return report, fmt.Errorf("entry stage not found: %s", entry)
	}

	runner := lp.makeStageRunner()
	for st != nil {
		// 检查是否已取消或超时
		if rc.Err() != nil {
			report.StageResults[st.Name()] = core.StageResult{
				Status: core.StageFailed,
				Err:    rc.Err(),
			}
			break
		}

		result := runner(rc, st)
		report.StageOrder = append(report.StageOrder, st.Name())
		report.StageResults[st.Name()] = result

		// 合并输出到共享状态
		for k, v := range result.Outputs {
			rc.Values[k] = v
		}

		if result.IsFailed() {
			report.Success = false
			break
		}

		st = nil // Linear 模式中，每个 stage 执行一次就结束
		report.Success = true
	}

	report.DurationMs = time.Since(start).Milliseconds()
	return report, nil
}

func (lp *LinearPipeline) makeStageRunner() func(*core.Context, core.Stage) core.StageResult {
	runner := func(rc *core.Context, st core.Stage) core.StageResult {
		return st.Run(rc)
	}

	// 从后往前包裹中间件
	for i := len(lp.mw) - 1; i >= 0; i-- {
		runner = lp.mw[i](runner)
	}

	return runner
}
