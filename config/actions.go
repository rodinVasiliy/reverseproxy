package config

import (
	"log"
)

type Action struct {
	name   string                 // название action
	action func(ap *ActionParams) //  то, что делает сам action
}

func LogToDB() func(*ActionParams) {
	return func(ap *ActionParams) {
		ip := ap.rp.ip
		rule := ap.rule
		log.Printf("ip: %s rule: %s", ip, rule)
	}
}

func BlockRequest() func(*ActionParams) {
	return func(ap *ActionParams) {
		// тут ничего не надо, этот action служит сигналом, что запрос надо блокировать на WAF/Proxy
	}
}

func (a *Action) Name() string {
	return a.name
}

func InitActions() []Action {
	initLogFile()
	return []Action{
		{
			name:   "Log to DB",
			action: LogToDB(),
		},
		{
			name:   "Block Request",
			action: BlockRequest(),
		},
	}
}

func (a *Action) DoAction(ap *ActionParams) {
	a.action(ap)
}
