package cond

import "github.com/KKKKKKKEM/flowkit/core"

// CondFunc 是条件判断函数，基于当前 Context 返回 true/false
type CondFunc func(rc *core.Context) bool

// Branch 是一个条件分支：满足 When 时跳转到 Next
type Branch struct {
	When CondFunc
	Next string
}

// Option 是 Stage 的配置项
type Option func(*Stage)

// WithBranch 添加一个条件分支，按添加顺序匹配，先添加的优先级高。
func WithBranch(when CondFunc, next string) Option {
	return func(s *Stage) {
		s.branches = append(s.branches, Branch{When: when, Next: next})
	}
}

// WithFallback 设置兜底跳转目标，所有分支都不命中时使用。
func WithFallback(next string) Option {
	return func(s *Stage) {
		s.fallback = next
	}
}
