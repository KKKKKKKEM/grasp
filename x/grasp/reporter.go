package grasp

import (
	"sync"
	"time"

	"github.com/KKKKKKKEM/flowkit/builtin/download"
	"github.com/KKKKKKKEM/flowkit/builtin/serve"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type ProgressReporter interface {
	Track(task *download.Task)
	Wait()
}

type MpbReporter struct {
	p *mpb.Progress
}

func NewMpbReporter() *MpbReporter {
	progress := mpb.New(mpb.WithRefreshRate(120 * time.Millisecond))

	return &MpbReporter{p: progress}
}

func (r *MpbReporter) Track(task *download.Task) {
	savePath, err := task.GetSavePath()
	if err != nil {
		savePath = task.Request.URL.String()
	}

	bar := r.p.AddBar(0,
		mpb.PrependDecorators(
			decor.Name(savePath+" ", decor.WCSyncWidth),
			decor.Counters(decor.SizeB1024(0), "% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.OnComplete(
				decor.EwmaETA(decor.ET_STYLE_GO, 30, decor.WCSyncWidth),
				"done",
			),
			decor.Name(" "),
			decor.EwmaSpeed(decor.SizeB1024(0), "% .2f", 30, decor.WCSyncWidth),
		),
	)

	var (
		mu         sync.Mutex
		knownTotal int64
		lastBytes  int64
	)

	origProgress := task.OnProgress
	task.OnProgress = func(downloaded, total int64) {
		mu.Lock()
		defer mu.Unlock()
		if total > 0 && total != knownTotal {
			knownTotal = total
			bar.SetTotal(total, false)
		}
		if lastBytes == 0 {
			lastBytes = downloaded
			bar.SetCurrent(downloaded)
		} else if delta := downloaded - lastBytes; delta > 0 {
			bar.EwmaIncrInt64(delta, 120*time.Millisecond)
			lastBytes = downloaded
		}
		if origProgress != nil {
			origProgress(downloaded, total)
		}
	}

	origComplete := task.OnComplete
	task.OnComplete = func(result *download.Result) {
		mu.Lock()
		bar.SetTotal(-1, true)
		mu.Unlock()
		if origComplete != nil {
			origComplete(result)
		}
	}
}

func (r *MpbReporter) Wait() {
	r.p.Wait()
}

type DownloadProgressData struct {
	URL        string `json:"url"`
	Downloaded int64  `json:"downloaded"`
	Total      int64  `json:"total"`
}

type DownloadCompleteData struct {
	URL  string `json:"url"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

type SSEReporter struct {
	sess *serve.SSESession
}

func NewSSEReporter(sess *serve.SSESession) *SSEReporter {
	return &SSEReporter{sess: sess}
}

func (r *SSEReporter) Track(task *download.Task) {
	url := task.Request.URL.String()

	origProgress := task.OnProgress
	task.OnProgress = func(downloaded, total int64) {
		r.sess.EmitProgress(DownloadProgressData{
			URL:        url,
			Downloaded: downloaded,
			Total:      total,
		})
		if origProgress != nil {
			origProgress(downloaded, total)
		}
	}

	origComplete := task.OnComplete
	task.OnComplete = func(result *download.Result) {
		r.sess.EmitProgress(DownloadCompleteData{
			URL:  url,
			Path: result.Path,
			Size: result.Size,
		})
		if origComplete != nil {
			origComplete(result)
		}
	}
}

func (r *SSEReporter) Wait() {}
