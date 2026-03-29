package extract

type stageOptions struct {
	fallback      Opts
	inputKey      string
	nextStageName string
	maxRounds     int
}

type Option func(*stageOptions)

func WithFallback(opts *Opts) Option {
	return func(o *stageOptions) {
		if opts != nil {
			o.fallback = *opts
		}
	}
}

func WithInputKey(inputKey string) Option {
	return func(o *stageOptions) { o.inputKey = inputKey }
}

func WithNextStage(stageName string) Option {
	return func(o *stageOptions) { o.nextStageName = stageName }
}

func WithMaxRounds(n int) Option {
	return func(o *stageOptions) { o.maxRounds = n }
}
