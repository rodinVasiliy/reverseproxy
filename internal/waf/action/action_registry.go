package action

import (
	"log"
	bl "reverseproxy/internal/infrastructure/config/bl"
	"time"
)

type Registry struct {
	actions map[string]ActionLogic
}

func NewActionRegistry(logger *log.Logger, bl bl.Blacklist) *Registry {
	return &Registry{
		actions: map[string]ActionLogic{
			LOG_TO_DB_ACTION_NAME:     &LogToDBAction{Logger: logger},
			SEND_TO_BL_ACTION_NAME:    &SendToBLAction{BlackList: bl, defaultTtl: time.Hour * 24},
			BLOCK_REQUEST_ACTION_NAME: &BlockRequestAction{},
		},
	}
}

func (r *Registry) Get(name string) (ActionLogic, bool) {
	a, ok := r.actions[name]
	return a, ok
}
