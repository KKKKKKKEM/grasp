package core

type TrackBuilder interface {
	Track(tag string, meta map[string]any) Tracker
	Done()
}

type Tracker interface {
	Set(key string, value any)
	Update(d map[string]any)
	Get(key string, fallback any) any
	Wait()
}

const trackBuilderKey = "__trackBuilder__"

func (rc *Context) WithTrackBuilder(builder TrackBuilder) {
	rc.Values[trackBuilderKey] = builder
}

func (rc *Context) TrackBuilder() TrackBuilder {
	builder, _ := rc.Values[trackBuilderKey].(TrackBuilder)
	return builder
}
