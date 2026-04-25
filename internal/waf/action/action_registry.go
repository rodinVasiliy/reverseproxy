package action

import (
	"log"
	"time"
)

type Factory func() Action

type Registry struct {
	factories map[string]Factory
}

func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]Factory),
	}
}

func (r *Registry) Register(name string, factory Factory) {
	r.factories[name] = factory
}

func (r *Registry) Get(name string) (Action, bool) {
	f, ok := r.factories[name]
	if !ok {
		return nil, false
	}
	return f(), true
}

func BuildRegistry(logger *log.Logger) *Registry {
	r := NewRegistry()

	r.Register(LogToDbActionName, func() Action {
		return &LogToDBAction{
			Logger: logger,
		}
	})

	r.Register(SendToBlActionName, func() Action {
		return &SendToBLAction{
			defaultTtl: 10 * time.Minute,
		}
	})

	r.Register(BlockRequestActionName, func() Action {
		return &BlockRequestAction{}
	})

	return r
}
