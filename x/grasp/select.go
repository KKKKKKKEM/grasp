package grasp

import (
	"context"

	"github.com/KKKKKKKEM/flowkit/x/download"
	"github.com/KKKKKKKEM/flowkit/x/extract"
)

type SelectFunc func(ctx context.Context, items []extract.ParseItem) ([]extract.ParseItem, error)

type TransformFunc func(ctx context.Context, item extract.ParseItem, baseOpts *download.Opts) (*download.Task, error)

func SelectAll(_ context.Context, items []extract.ParseItem) ([]extract.ParseItem, error) {
	return items, nil
}

func SelectFirst(n int) SelectFunc {
	return func(_ context.Context, items []extract.ParseItem) ([]extract.ParseItem, error) {
		if n >= len(items) {
			return items, nil
		}
		return items[:n], nil
	}
}

func SelectByIndices(indices []int) SelectFunc {
	return func(_ context.Context, items []extract.ParseItem) ([]extract.ParseItem, error) {
		var out []extract.ParseItem
		for _, idx := range indices {
			if idx >= 0 && idx < len(items) {
				out = append(out, items[idx])
			}
		}
		return out, nil
	}
}

func DefaultTransform(baseOpts *download.Opts) TransformFunc {
	return func(_ context.Context, item extract.ParseItem, opts *download.Opts) (*download.Task, error) {
		if opts == nil {
			opts = baseOpts
		}
		if opts == nil {
			opts = &download.Opts{}
		}
		headers, _ := item.Meta["headers"].(map[string]string)
		return download.NewTaskFromURI(item.URI, opts, headers)
	}
}
