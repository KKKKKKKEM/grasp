package core

import "fmt"

type Registry struct {
	downloaders []Downloader
}

func NewRegistry() *Registry {
	return &Registry{}
}

// Register 注册顺序即优先级：先注册的优先匹配，通用兜底最后注册。
func (r *Registry) Register(d Downloader) {
	r.downloaders = append(r.downloaders, d)
}

func (r *Registry) Resolve(task *DownloadTask) (Downloader, error) {
	for _, d := range r.downloaders {
		if d.CanHandle(task) {
			return d, nil
		}
	}
	return nil, &ErrNoDownloader{URL: task.URL}
}

type ErrNoDownloader struct {
	URL string
}

func (e *ErrNoDownloader) Error() string {
	return fmt.Sprintf("no downloader for URL: %s", e.URL)
}
