package log

type Option func(a *Adapter)

func WithDevelopmentMode() Option {
	return func(a *Adapter) {
		a.devMode = true
	}
}

func WithFileRotation(r Rotation) Option {
	return func(a *Adapter) {
		a.rotation = &r
	}
}

func WithLevel(level Level) Option {
	return func(a *Adapter) {
		a.initialLevel = level
	}
}
