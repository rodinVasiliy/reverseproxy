package initialization

// TODO

import (
	"context"
	"fmt"
	actiondoc "reverseproxy/internal/domain/action"
	action "reverseproxy/internal/waf/action"
)

// поянение к action и actiondoc:
// actionDoc - лежит в Mongo, содержит только имя действия, используется в политиках / правилах
// action - реализует конкретную логику: block | log

func getDefaultActions() []actiondoc.ActionDoc {
	var actionDocs []actiondoc.ActionDoc
	logToDBActionDoc := actiondoc.ActionDoc{
		Name: action.LogToDbActionName,
	}
	blockRequestActionDoc := actiondoc.ActionDoc{
		Name: action.BlockRequestActionName,
	}
	sendToBLActionDoc := actiondoc.ActionDoc{
		Name: action.SendToBlActionName,
	}
	actionDocs = append(actionDocs, logToDBActionDoc)
	actionDocs = append(actionDocs, blockRequestActionDoc)
	actionDocs = append(actionDocs, sendToBLActionDoc)
	return actionDocs
}

func loadActionsToDB(as *actiondoc.Service, actions []actiondoc.ActionDoc) error {
	ctx := context.Background()
	for _, action := range actions {
		_, err := as.Insert(ctx, action)
		if err != nil {
			return fmt.Errorf("failed to add action %s to db %w", action.Name, err)
		}
	}
	return nil
}
