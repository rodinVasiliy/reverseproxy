package action

import (
	"log"
	bl "reverseproxy/internal/infrastructure/config/bl"
	parsedRequest "reverseproxy/internal/waf/parsed_request"
	"time"
)

var (
	LOG_TO_DB_ACTION_NAME     = "Log to DB"
	BLOCK_REQUEST_ACTION_NAME = "Block Requst"
	SEND_TO_BL_ACTION_NAME    = "Send to BL"
)

type ActionParams struct {
	Rule string                       // название правила, под которое запрос попал
	PR   *parsedRequest.ParsedRequest // сам запрос
}

type ActionLogic interface {
	Do(ap *ActionParams) error
}

type LogToDBAction struct {
	Logger *log.Logger
}

func (a *LogToDBAction) Do(ap *ActionParams) error {
	a.Logger.Printf("IP=%s Rule=%s", ap.PR.IP, ap.Rule)
	return nil
}

type SendToBLAction struct {
	BlackList  bl.Blacklist
	defaultTtl time.Duration
}

func (a *SendToBLAction) Do(ap *ActionParams) error {
	return a.BlackList.Add(ap.PR.IP.String(), a.defaultTtl)
}

type BlockRequestAction struct {
}

// это просто индикатор того, что запрос нужно заблокировать
func (a *BlockRequestAction) Do(ap *ActionParams) error {
	return nil
}
