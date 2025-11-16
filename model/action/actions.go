package action

import (
	"log"
	parsedRequest "reverseproxy/model/parsed_request"
)

type ActionType int

const (
	ActionNormal ActionType = iota // уникальный акшион который выполняется 1 раз для запроса(отправка в BL, FW и прочие радости)
	ActionLog
	ActionBlock
)

type Action struct {
	Name       string                 // название action
	Do         func(ap *ActionParams) //  то, что делает сам action
	ActionType ActionType
}

type ActionParams struct {
	rule string
	rp   *parsedRequest.ParsedRequest
}

// Сами Actions

func LogToDB() func(*ActionParams) {
	return func(ap *ActionParams) {
		ip := ap.rp.IP
		rule := ap.rule
		log.Printf("ip: %s rule: %s", ip, rule)
	}
}

func BlockRequest() func(*ActionParams) {
	return func(ap *ActionParams) {
		// тут ничего не надо, этот action служит сигналом, что запрос надо блокировать на WAF/Proxy
	}
}

func ActionsByName(names ...string) []*Action {
	var result []*Action
	for _, name := range names {
		result = append(result, namesAndActionsMap[name])
	}
	return result
}

var (
	LOG_TO_DB_ACTION_NAME     = "Log to DB"
	BLOCK_REQUEST_ACTION_NAME = "BlockRequst"

	logToDBAction = &Action{
		Name:       "Log to DB",
		Do:         LogToDB(),
		ActionType: ActionLog,
	}
	blockRequestAction = &Action{
		Name:       "Block request",
		Do:         BlockRequest(),
		ActionType: ActionBlock,
	}

	namesAndActionsMap = map[string]*Action{
		LOG_TO_DB_ACTION_NAME:     logToDBAction,
		BLOCK_REQUEST_ACTION_NAME: blockRequestAction,
	}
)
