package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func initLogFile(port int) (*os.File, error) {
	logFileName := filepath.Join("log", fmt.Sprintf("db_%d.log", port))
	var err error
	var logFile *os.File
	logFile, err = os.OpenFile(logFileName,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	// Настраиваем логгер на запись в файл
	log.SetOutput(logFile)
	return logFile, err
}

// TODO подумать, почему я вызываю это в main и делаю метод открытым...
func (cfg *Config) CloseLogFile() {
	if cfg.logFile != nil {
		cfg.logFile.Close()
		fmt.Println("log file closed")
	}
}
