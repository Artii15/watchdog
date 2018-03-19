package loggers

import (
	"log"
	"io"
	"os"
	"strings"
	"io/ioutil"
	"os/user"
)

const (
	errorPrefix   = "ERROR: "
	warningPrefix = "WARNING: "
	infoPrefix    = "INFO: "
)

const logfileBaseName = "watchdog.log"


type Logs struct {
	config Config
	messagesChannel chan Message
	currentLogFileSize int
	logger *log.Logger
	currentLogfile *os.File
}

func (logs *Logs) Close() {
	close(logs.messagesChannel)
}

func (logs *Logs) Error(message string)  {
	logs.messagesChannel <- Message{Prefix: errorPrefix, Content: message}
}

func (logs *Logs) Info(message string)  {
	logs.messagesChannel <- Message{Prefix: infoPrefix, Content: message}
}

func (logs *Logs) Warning(message string)  {
	logs.messagesChannel <- Message{Prefix: warningPrefix, Content: message}
}

func New(config Config) (*Logs, error) {
	var logs Logs
	logs.config = config
	logs.messagesChannel = make(chan Message)
	logs.currentLogFileSize = 0

	logger, currentLogfile, err := setupLogger(config.logsDirPath)
	if err != nil {
		return nil, err
	}
	logs.logger = logger
	logs.currentLogfile = currentLogfile

	go logs.runWorker()

	return &logs, nil
}

func setupLogger(logsDirPath string) (*log.Logger, *os.File, error) {
	logfile, err := openLogfile(logsDirPath)
	if err != nil {
		return nil, nil, err
	}
	return createLogger(logfile, ""), logfile, nil
}

func openLogfile(directoryPath string) (*os.File, error) {
	normalizedDirPath := strings.TrimRight(directoryPath, "/")
	filePath := strings.Join([]string{normalizedDirPath, logfileBaseName}, "/")
	return os.OpenFile(filePath, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0644)
}

func createLogger(writer io.Writer, prefix string) *log.Logger {
	return log.New(writer, prefix, log.Ldate|log.Ltime)
}

func (logs *Logs) runWorker() {
	defer logs.closeLogfile()

	for message := range logs.messagesChannel {
		logs.log(message)
	}
}

func (logs *Logs) closeLogfile() {
	if logs.currentLogfile != nil {
		logs.currentLogfile.Close()
	}
}

func (logs *Logs) log(message Message)  {
	logs.currentLogFileSize = logs.updatedLogFileSize(message.Content)
	logs.changeLogfileIfTooBig(logs.currentLogFileSize)

	logs.logger.SetPrefix(message.Prefix)
	logs.logger.Println(message.Content)
}

func (logs *Logs) updatedLogFileSize(message string) int {
	return logs.currentLogFileSize + len(message)
}

func (logs *Logs) changeLogfileIfTooBig(fileSize int)  {
	if logs.currentLogFileSize > logs.config.logFileMaxSize {
		//TODO send file to S3 asynchronously
		logCopy, err := ioutil.TempFile(logs.config.logsDirPath, "logs")
		if err != nil {
			//TODO handle send file error
		}
		_, err = io.Copy(logCopy, logs.currentLogfile)
		if err != nil {
			//TODO handle copy creation error
		}
		go sendToS3(logCopy)

		logs.currentLogfile.Truncate(0)
		logs.currentLogFileSize = 0
	}
}

func sendToS3(file *os.File) {
	//TODO
}
