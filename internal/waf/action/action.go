package action

import (
	"log"
	bl "reverseproxy/internal/infrastructure/config/bl"
	parsedrequest "reverseproxy/internal/waf/parsedrequest"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	LogToDbActionName      = "Log to DB"
	BlockRequestActionName = "Block Request"
	SendToBlActionName     = "Send to BL"
)

type Action interface {
	Execute(context *Context) Effect
}

type Context struct {
	Request *parsedrequest.ParsedRequest
	RuleId  primitive.ObjectID
}

/////////////////////////// Реализации ///////////////////////////

type LogToDBAction struct {
	Logger *log.Logger
}

func (a *LogToDBAction) Execute(ctx *Context) Effect {
	return Effect{
		Logs: []LogEntry{
			{
				Message: "request matched rule",
				Fields: map[string]string{
					"ip":   ctx.Request.IP.String(),
					"rule": ctx.RuleId.Hex(),
				},
			},
		},
	}
}

type SendToBLAction struct {
	defaultTtl time.Duration
}

func (a *SendToBLAction) Execute(ctx *Context) Effect {
	return Effect{
		AddToBL: &BLRequest{
			IP:  ctx.Request.IP.String(),
			TTL: a.defaultTtl,
		},
	}
}

type BlockRequestAction struct {
}

func (a *BlockRequestAction) Execute(ctx *Context) Effect {
	return Effect{
		Block: true,
	}
}

func ExecuteActions(actions []Action, ctx *Context) Effect {
	var result Effect
	for _, action := range actions {
		eff := action.Execute(ctx)

		if eff.Block {
			result.Block = true
		}

		result.Logs = append(result.Logs, eff.Logs...)

		if eff.AddToBL != nil {
			result.AddToBL = eff.AddToBL
		}
	}
	return result
}

func ApplyEffects(eff Effect, logger *log.Logger, bl *bl.RedisBL) error {
	for _, logEntry := range eff.Logs {
		logger.Println(logEntry.Message, logEntry.Fields)
	}

	if eff.AddToBL != nil {
		err := bl.Add(eff.AddToBL.IP, eff.AddToBL.TTL)
		if err != nil {
			return err
		}
	}
	return nil
}
