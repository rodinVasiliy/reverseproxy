package log

import (
	"fmt"
	"os"
)

type LogConfig struct {
	file *os.File // файл куда будет записываться лог
	path string   // путь к файлу с логами
}

func (c *LogConfig) File() *os.File {
	return c.file
}

// создает файл для лога, по пути, указанному в параметре logFileName
func NewLogConfig(logFileName string) (*LogConfig, error) {
	var err error
	var logFile *os.File
	logFile, err = os.OpenFile(logFileName,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &LogConfig{file: logFile, path: logFileName}, err
}

// TO DO подумать, почему я вызываю это в main и делаю метод открытым...
func (logConfig *LogConfig) CloseLogFile() {
	if logConfig.file != nil {
		logConfig.file.Close()
		fmt.Println("log file closed")
	}
}
