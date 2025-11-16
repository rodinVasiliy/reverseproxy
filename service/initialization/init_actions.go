package initialization

// TODO

import (
	"fmt"
	action "reverseproxy/model/action"
	service "reverseproxy/service"
)

// func GetDefaultActions() *[]action.ActionDoc {

// }

func LoadActionsToDB(as *service.ActionService) error {

	logToDBActionDoc := action.ActionDoc{
		Name: LOG_TO_DB_ACTION_NAME,
	}
	blockRequestActionDoc := action.ActionDoc{
		Name: BLOCK_REQUEST_ACTION_NAME,
	}

	_, err := as.Add(&logToDBActionDoc)
	if err != nil {
		return fmt.Errorf("failed to add action doc to db %w", err)
	}
	_, err = as.Add(&blockRequestActionDoc)
	if err != nil {
		return fmt.Errorf("failed to add action doc to db %w", err)
	}
	return nil
}

var (
	LOG_TO_DB_ACTION_NAME     = "Log to DB"
	BLOCK_REQUEST_ACTION_NAME = "BlockRequst"

	// logToDBAction = &Action{
	// 	Name:       "Log to DB",
	// 	Do:         LogToDB(),
	// 	ActionType: ActionLog,
	// }
	// blockRequestAction = &Action{
	// 	Name:       "Block request",
	// 	Do:         BlockRequest(),
	// 	ActionType: ActionBlock,
	// }

	// namesAndActionsMap = map[string]*Action{
	// 	logToDBActionName:      logToDBAction,
	// 	blockRequestActionName: blockRequestAction,
	// }
)
