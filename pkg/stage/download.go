package stage

import (
	"context"

	"github.com/KKKKKKKEM/grasp/pkg/core"
)

type ResolveStage struct {
	Task     *core.DownloadTask
	Registry *core.Registry
}

func (s *ResolveStage) Do(ctx context.Context) (core.Stage, error) {
	dl, err := s.Registry.Resolve(s.Task)
	if err != nil {
		return nil, err
	}
	return &DownloadStage{Task: s.Task, Downloader: dl}, nil
}

type DownloadStage struct {
	Task       *core.DownloadTask
	Downloader core.Downloader
}

func (s *DownloadStage) Do(ctx context.Context) (core.Stage, error) {
	result, err := s.Downloader.Download(ctx, s.Task)
	if err != nil {
		return nil, err
	}
	if s.Task.OnComplete != nil {
		s.Task.OnComplete(result)
	}
	return nil, nil
}
