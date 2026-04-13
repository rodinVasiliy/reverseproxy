package action

import (
	"log"
	bl "reverseproxy/internal/infrastructure/config/bl"
	parsedRequest "reverseproxy/internal/waf/parsed_request"
	"time"
)

var (
	LogToDbActionName      = "Log to DB"
	BlockRequestActionName = "Block Requst"
	SendToBlActionName     = "Send to BL"
)

type Params struct {
	Rule string                       // название правила, под которое запрос попал
	PR   *parsedRequest.ParsedRequest // сам запрос
}

type Logic interface {
	Do(ap *Params) error
}

type LogToDBAction struct {
	Logger *log.Logger
}

func (a *LogToDBAction) Do(ap *Params) error {
	a.Logger.Printf("IP=%s Rule=%s", ap.PR.IP, ap.Rule)
	return nil
}

type SendToBLAction struct {
	BlackList  bl.Blacklist
	defaultTtl time.Duration
}

func (a *SendToBLAction) Do(ap *Params) error {
	return a.BlackList.Add(ap.PR.IP.String(), a.defaultTtl)
}

type BlockRequestAction struct {
}

// это просто индикатор того, что запрос нужно заблокировать
func (a *BlockRequestAction) Do(ap *Params) error {
	return nil
}
