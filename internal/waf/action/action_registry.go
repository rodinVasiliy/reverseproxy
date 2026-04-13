package action

import (
	"log"
	bl "reverseproxy/internal/infrastructure/config/bl"
	"time"
)

type Registry struct {
	actions map[string]Logic
}

func NewActionRegistry(logger *log.Logger, bl bl.Blacklist) *Registry {
	return &Registry{
		actions: map[string]Logic{
			LogToDbActionName:      &LogToDBAction{Logger: logger},
			SendToBlActionName:     &SendToBLAction{BlackList: bl, defaultTtl: time.Hour * 24},
			BlockRequestActionName: &BlockRequestAction{},
		},
	}
}

func (r *Registry) Get(name string) (Logic, bool) {
	a, ok := r.actions[name]
	return a, ok
}
