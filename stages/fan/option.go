package fan

import "fmt"

type WaitStrategy int

const (
	WaitAll WaitStrategy = iota
	WaitAny
)

type FailStrategy int

const (
	FailFast FailStrategy = iota
	BestEffort
)

type ConflictStrategy int

const (
	OverwriteOnConflict ConflictStrategy = iota
	ErrorOnConflict
)

type options struct {
	wait     WaitStrategy
	fail     FailStrategy
	conflict ConflictStrategy
}

func defaultOptions() options {
	return options{
		wait:     WaitAll,
		fail:     FailFast,
		conflict: OverwriteOnConflict,
	}
}

type Option func(*options)

func WithWaitStrategy(s WaitStrategy) Option {
	return func(o *options) { o.wait = s }
}

func WithFailStrategy(s FailStrategy) Option {
	return func(o *options) { o.fail = s }
}

func WithConflictStrategy(s ConflictStrategy) Option {
	return func(o *options) { o.conflict = s }
}

func mergeOutputs(dst map[string]any, src map[string]any, strategy ConflictStrategy) error {
	for k, v := range src {
		if _, exists := dst[k]; exists && strategy == ErrorOnConflict {
			return fmt.Errorf("output key conflict: %q", k)
		}
		dst[k] = v
	}
	return nil
}
