package core

import (
	"context"
	"io"
)

type DownloadTask struct {
	URL         string
	Dest        string
	Headers     map[string]string
	Concurrency int
	OnProgress  ProgressFunc
	OnComplete  CompleteFunc
	Meta        map[string]any
}

// CompleteFunc 在下载成功后调用，result 包含实际文件路径和写入字节数。
type CompleteFunc func(result *DownloadResult)

// ProgressFunc 报告下载进度；total 为 -1 表示总大小未知。
type ProgressFunc func(downloaded, total int64)

type DownloadResult struct {
	FilePath     string
	BytesWritten int64
}

type Downloader interface {
	// CanHandle 根据 task 的任意字段综合判断，不应发起网络请求。
	CanHandle(task *DownloadTask) bool
	Download(ctx context.Context, task *DownloadTask) (*DownloadResult, error)
	Name() string
}

// StreamDownloader 是可选扩展，适用于不落盘直接消费字节流的场景。
type StreamDownloader interface {
	Downloader
	Stream(ctx context.Context, task *DownloadTask) (io.ReadCloser, error)
}
