package core

import "context"

// Stage 是状态机中的单个节点。
type Stage interface {
	Do(ctx context.Context) (next Stage, err error)
}

// Pipeline 是状态机的驱动器。
type Pipeline interface {
	Run(ctx context.Context, stage Stage) error
}

type pipeline struct{}

func NewPipeline() Pipeline {
	return &pipeline{}
}

func (p *pipeline) Run(ctx context.Context, stage Stage) error {
	for stage != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
		next, err := stage.Do(ctx)
		if err != nil {
			return err
		}
		stage = next
	}
	return nil
}
