package download

type stageOptions struct {
	fallback      Opts
	headers       map[string]string
	inputKey      string
	nextStageName string
}

type Option func(*stageOptions)

func WithFallback(opts *Opts) Option {
	return func(o *stageOptions) {
		if opts != nil {
			o.fallback = *opts
		}
	}
}

func WithFallbackHeaders(headers map[string]string) Option {
	return func(o *stageOptions) { o.headers = headers }
}

func WithInputKey(inputKey string) Option {
	return func(o *stageOptions) { o.inputKey = inputKey }
}

func WithNextStage(stageName string) Option {
	return func(o *stageOptions) { o.nextStageName = stageName }
}
