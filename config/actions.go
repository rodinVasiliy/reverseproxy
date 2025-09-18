package config

import (
	"fmt"
	"log"
	"os"
)

var logFileName = "log\\db.log"

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

// to do add action send to BL

func initLogFile() {
	file, err := os.OpenFile(logFileName,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v", err)
	}
	defer file.Close()

	// Настраиваем логгер на запись в файл
	log.SetOutput(file)
}

// to do add action block request

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
