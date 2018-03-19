package loggers

import (
	"log"
	"io"
	"os"
	"strings"
	"io/ioutil"
)

const ErrorPrefix  = "ERROR: "
const WarningPrefix  = "WARNING: "
const InfoPrefix  = "INFO: "

const logfileBaseName = "watchdog.log"

type Logs struct {
	config Config
	messagesChannel <-chan Message
	usingStdOut bool
	currentLogfile *os.File
	logger *log.Logger
	currentLogFileSize int
}

func New(config Config, messagesChannel <-chan Message) *Logs {
	var logs Logs
	logs.config = config
	logs.messagesChannel = messagesChannel
	logs.usingStdOut = true
	logs.currentLogFileSize = 0
	go logs.runWorker()

	return &logs
}

func (logs *Logs) runWorker()  {
	logs.setupLogger()
	logs.warnIfUsingStdout()
	defer logs.closeLogfile()

	for message := range logs.messagesChannel {

	}
}

func (logs *Logs) setupLogger() *log.Logger {
	var logWriter io.Writer
	logfile, err := openLogfile(logs.config.LogsDirPath)
	if err == nil {
		logs.currentLogfile = logfile
		logWriter = logfile
		logs.usingStdOut = false
	} else {
		logWriter = os.Stdout
		logs.usingStdOut = true
	}
	logs.logger = createLogger(logWriter, "")
}

func (logs *Logs) warnIfUsingStdout() {
	if logs.usingStdOut {
		logs.logMessage(Message{Prefix: ErrorPrefix, Content: "Could not open logfile, fallback to stdout"})
	}
}

func createLogger(writer io.Writer, prefix string) *log.Logger {
	return log.New(writer, prefix, log.Ldate|log.Ltime)
}

func (logs *Logs) closeLogfile() {
	if logs.currentLogfile != nil {
		logs.currentLogfile.Close()
	}
}

func openLogfile(directoryPath string) (*os.File, error) {
	normalizedDirPath := strings.TrimRight(directoryPath, "/")
	filePath := strings.Join([]string{normalizedDirPath, logfileBaseName}, "/")
	return os.OpenFile(filePath, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0644)
}

func (logs *Logs) logMessage(message Message)  {
	logs.updateLogFileSize(message.Content)

	logs.logger.SetPrefix(message.Prefix)
	logs.logger.Println(message.Content)
}

func (logs *Logs) updateLogFileSize(message string)  {
	if !logs.usingStdOut {
		logs.currentLogFileSize = logs.currentLogFileSize + len(message)
	} else {
		logs.currentLogFileSize = 0
	}
}

func (logs *Logs) changeLogFileIfTooBig()  {
	if !logs.usingStdOut && logs.currentLogFileSize > logs.config.LogFileMaxSize {
		//TODO send file to S3 asynchronously
		logCopy, err := ioutil.TempFile(logs.config.LogsDirPath, "logs")
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
