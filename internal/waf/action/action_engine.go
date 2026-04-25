package action

import (
	"fmt"
	"reverseproxy/internal/domain/action"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ActionEngine struct {
	actions map[primitive.ObjectID]Action
}

func (ae *ActionEngine) Load(actions []action.ActionDoc, registry *Registry) error {
	actionMap := make(map[primitive.ObjectID]Action)
	for _, a := range actions {
		act, ok := registry.Get(a.Name)
		if !ok {
			return fmt.Errorf("action %s not found", a.Name)
		}
		actionMap[a.ID] = act
	}
	ae.actions = actionMap
	return nil
}

func (ae *ActionEngine) Get(id primitive.ObjectID) (Action, bool) {
	act, ok := ae.actions[id]
	return act, ok
}
