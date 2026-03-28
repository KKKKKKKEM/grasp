package grasp

import (
	"sync"
	"time"

	"github.com/KKKKKKKEM/flowkit/core"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type MPBTrackerProvider struct {
	progress *mpb.Progress
}

func (r *MPBTrackerProvider) Track(tag string, meta map[string]any) core.Tracker {
	total, ok := meta["total"]
	if !ok {
		total = int64(0)
	} else {
		if t, ok := total.(int64); ok {
			total = t
		} else {
			total = int64(0)
		}
	}

	meta["tag"] = tag
	meta["total"] = total

	bar := r.progress.AddBar(total.(int64),
		mpb.PrependDecorators(
			decor.Name(tag+" ", decor.WCSyncWidth),
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

	return &MpbTracker{bar: bar, progress: r.progress, meta: meta}
}

func (r *MPBTrackerProvider) Wait() {
	r.progress.Wait()
}

type MpbTracker struct {
	progress *mpb.Progress
	mu       sync.Mutex
	bar      *mpb.Bar
	meta     map[string]any
}

func (t *MpbTracker) Update(d map[string]any) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for key, value := range d {
		if key == "current" {
			hisCurrent, ok := t.meta["current"].(int64)
			if !ok {
				hisCurrent = int64(0)
			}
			newCurrent, ok := value.(int64)
			if !ok {
				newCurrent = int64(0)
			}
			delta := newCurrent - hisCurrent
			if delta > 0 {
				t.meta["delta"] = delta
			}
		}
		t.meta[key] = value
	}

}

func (t *MpbTracker) Flush() {
	t.mu.Lock()
	defer t.mu.Unlock()

	var (
		current int64
		total   int64
		delta   int64
		ok      bool
	)

	if _, ok = t.meta["current"]; ok {
		current, ok = t.meta["current"].(int64)
		if !ok {
			current = int64(0)
		}
	} else {
		current = int64(0)
	}

	if _, ok = t.meta["delta"]; ok {
		delta, ok = t.meta["delta"].(int64)
		if !ok {
			delta = int64(0)
		}
		t.bar.IncrInt64(delta)

	} else {
		t.bar.SetCurrent(current)
	}

	if _, ok = t.meta["total"]; ok {
		total, ok = t.meta["total"].(int64)
		if ok {
			t.bar.SetTotal(total, false)
		}
	}

}

func (t *MpbTracker) Done() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.bar.SetTotal(-1, true)
}

func NewMPBTrackerProvider() *MPBTrackerProvider {
	return &MPBTrackerProvider{
		progress: mpb.New(mpb.WithRefreshRate(120 * time.Millisecond)),
	}
}
