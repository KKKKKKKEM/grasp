package cond

import "github.com/KKKKKKKEM/flowkit/core"

// Stage 条件跳转 Stage，按顺序匹配 Branch，第一个满足条件的分支胜出。
// 若没有分支命中则跳转到 Fallback；Fallback 为空表示流程结束。
// 适用于 FSMPipeline。
type Stage struct {
	name     string
	branches []Branch
	fallback string
}

// New 创建一个条件跳转 Stage。
func New(name string, opts ...Option) *Stage {
	s := &Stage{name: name}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Stage) Name() string { return s.name }

func (s *Stage) Run(rc *core.RunContext) core.StageResult {
	for _, b := range s.branches {
		if b.When(rc) {
			return core.StageResult{
				Status: core.StageSuccess,
				Next:   b.Next,
			}
		}
	}
	return core.StageResult{
		Status: core.StageSuccess,
		Next:   s.fallback,
	}
}
