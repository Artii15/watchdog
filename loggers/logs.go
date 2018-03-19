package loggers

import (
	"log"
	"io"
	"os"
	"strings"
	"io/ioutil"
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
	currentLogfile *os.File
	logger *log.Logger
	currentLogFileSize int
}

func New(config Config) *Logs {
	var logs Logs
	logs.config = config
	logs.messagesChannel = make(chan Message)
	logs.currentLogFileSize = 0
	logs.setupLogger()

	go logs.runWorker()

	return &logs
}

func (logs *Logs) setupLogger() *log.Logger {
	var logWriter io.Writer
	logfile, err := openLogfile(logs.config.logsDirPath)
	if err == nil {
		logs.currentLogfile = logfile
		logWriter = logfile
	}
	// todo return an error
	logs.logger = createLogger(logWriter, "")
}

func openLogfile(directoryPath string) (*os.File, error) {
	normalizedDirPath := strings.TrimRight(directoryPath, "/")
	filePath := strings.Join([]string{normalizedDirPath, logfileBaseName}, "/")
	return os.OpenFile(filePath, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0644)
}

func createLogger(writer io.Writer, prefix string) *log.Logger {
	return log.New(writer, prefix, log.Ldate|log.Ltime)
}

func (logs *Logs) runWorker()  {
	defer logs.closeLogfile()

	for message := range logs.messagesChannel {

	}
}

func (logs *Logs) closeLogfile() {
	if logs.currentLogfile != nil {
		logs.currentLogfile.Close()
	}
}

func (logs *Logs) logMessage(message Message)  {
	logs.updateLogFileSize(message.Content)

	logs.logger.SetPrefix(message.Prefix)
	logs.logger.Println(message.Content)
}

func (logs *Logs) updateLogFileSize(message string)  {
	logs.currentLogFileSize = logs.currentLogFileSize + len(message)
}

func (logs *Logs) changeLogFileIfTooBig()  {
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
