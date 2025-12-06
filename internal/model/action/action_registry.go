package action

import (
	"log"
	bl "reverseproxy/config/bl"
)

type Registry struct {
	actions map[string]ActionLogic
}

func NewActionRegistry(logger *log.Logger, bl *bl.BL) *Registry {
	return &Registry{
		actions: map[string]ActionLogic{
			LOG_TO_DB_ACTION_NAME:     &LogToDBAction{Logger: logger},
			SEND_TO_BL_ACTION_NAME:    &SendToBLAction{BlackList: bl},
			BLOCK_REQUEST_ACTION_NAME: &BlockRequestAction{},
		},
	}
}

func (r *Registry) Get(name string) (ActionLogic, bool) {
	a, ok := r.actions[name]
	return a, ok
}
