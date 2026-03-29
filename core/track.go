package core

type TrackerProvider interface {
	Track(tag string, meta map[string]any) Tracker
	Wait()
}

type Tracker interface {
	Update(d map[string]any)
	Flush()
	Done()
}

const trackerBuilderKey = "__trackerBuilder__"

func (rc *Context) WithTrackerProvider(builder TrackerProvider) {
	rc.Set(trackerBuilderKey, builder)
}

func (rc *Context) TrackerProvider() TrackerProvider {
	builder, _ := rc.Get(trackerBuilderKey)
	provider, _ := builder.(TrackerProvider)
	return provider
}
