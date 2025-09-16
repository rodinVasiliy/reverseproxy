package config

import (
	"log"
	"net/http"
	"os"
)

var logFileName = "db.log"

type Action struct {
	name   string                                   // название action
	rule   string                                   // название правила
	action func(http.ResponseWriter, *http.Request) // то, что делает сам action
}

func LogToDB(rule string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getIpFromRequest(r)
		log.Printf("ip: %s rule: %s", ip, rule)
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
		log.Fatalf("Failed to open file: %v", err)
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
			rule:   "",
			action: LogToDB(""),
		},
	}
}

func (a *Action) DoAction(w http.ResponseWriter, r *http.Request) {
	a.action(w, r)
}
