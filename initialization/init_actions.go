package initialization

// TODO

import (
	"context"
	"fmt"
	action "reverseproxy/internal/model/action"
)

func getDefaultActions() []action.ActionDoc {
	var actionDocs []action.ActionDoc
	logToDBActionDoc := action.ActionDoc{
		Name: action.LOG_TO_DB_ACTION_NAME,
	}
	blockRequestActionDoc := action.ActionDoc{
		Name: action.BLOCK_REQUEST_ACTION_NAME,
	}
	sendToBLActionDoc := action.ActionDoc{
		Name: action.SEND_TO_BL_ACTION_NAME,
	}
	actionDocs = append(actionDocs, logToDBActionDoc)
	actionDocs = append(actionDocs, blockRequestActionDoc)
	actionDocs = append(actionDocs, sendToBLActionDoc)
	return actionDocs
}

func loadActionsToDB(as *action.Service, actions []action.ActionDoc) error {
	ctx := context.Background()
	for _, action := range actions {
		_, err := as.Insert(ctx, action)
		if err != nil {
			return fmt.Errorf("failed to add action %s to db %w", action.Name, err)
		}
	}
	return nil
}
