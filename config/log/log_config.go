package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type LogConfig struct {
	file *os.File
}

// создает файл для лога, который потом нужно будет закрыть, в пути к файлу используется порт, который использует прокси
func NewLogConfig(port int) (*LogConfig, error) {
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
	return &LogConfig{file: logFile}, err
}

// TODO подумать, почему я вызываю это в main и делаю метод открытым...
func (logConfig *LogConfig) CloseLogFile() {
	if logConfig.file != nil {
		logConfig.file.Close()
		fmt.Println("log file closed")
	}
}
