package action

import (
	"log"
	bl "reverseproxy/config/bl"
	parsedRequest "reverseproxy/internal/model/parsed_request"
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
	BlackList *bl.BL
}

func (a *SendToBLAction) Do(ap *ActionParams) error {
	return a.BlackList.Add(ap.PR.IP.String())
}

type BlockRequestAction struct {
}

// это просто индикатор того, что запрос нужно заблокировать
func (a *BlockRequestAction) Do(ap *ActionParams) error {
	return nil
}
