package initialization

// TODO

import (
	"fmt"
	action "reverseproxy/model/action"
	service "reverseproxy/service"
)

func GetDefaultActions() *[]action.ActionDoc {
	var actionDocs []action.ActionDoc
	logToDBActionDoc := action.ActionDoc{
		Name: LOG_TO_DB_ACTION_NAME,
	}
	blockRequestActionDoc := action.ActionDoc{
		Name: BLOCK_REQUEST_ACTION_NAME,
	}
	actionDocs = append(actionDocs, logToDBActionDoc)
	actionDocs = append(actionDocs, blockRequestActionDoc)
	return &actionDocs
}

func LoadActionsToDB(as *service.ActionService, actions *[]action.ActionDoc) error {
	for _, action := range *actions {
		_, err := as.Add(&action)
		if err != nil {
			return fmt.Errorf("failed to add action %s to db %w", action.Name, err)
		}
	}
	return nil
}

var (
	LOG_TO_DB_ACTION_NAME     = "Log to DB"
	BLOCK_REQUEST_ACTION_NAME = "BlockRequst"
)
