package grasp

import (
	"github.com/KKKKKKKEM/flowkit/builtin/download"
	"github.com/KKKKKKKEM/flowkit/builtin/extract"
)

type Option func(*Pipeline)

func WithExtractor(e *extract.Stage) Option {
	return func(p *Pipeline) {
		p.extractor = e
	}
}

func WithDownloader(d *download.DirectDownloadStage) Option {
	return func(p *Pipeline) {
		p.downloader = d
	}
}

// WithSelector sets a pipeline-level default; Task.Selector takes precedence.
func WithSelector(fn SelectFunc) Option {
	return func(p *Pipeline) {
		p.defaultSelector = fn
	}
}

// WithTransform sets a pipeline-level default; Task.Transform takes precedence.
func WithTransform(fn TransformFunc) Option {
	return func(p *Pipeline) {
		p.defaultTransform = fn
	}
}

func WithProgress(r ProgressReporter) Option {
	return func(p *Pipeline) {
		p.reporter = r
	}
}
