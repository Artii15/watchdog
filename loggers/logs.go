package loggers

import (
	"log"
	"io"
	"os"
	"strings"
	"io/ioutil"
	"regexp"
	"fmt"
	"strconv"
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
	currentLogFileSize int64
	logger *log.Logger
	currentLogfile *os.File
}

func (logs *Logs) Close() {
	close(logs.messagesChannel)
}

func (logs *Logs) Error(messages ...interface{})  {
	logs.messagesChannel <- Message{Prefix: errorPrefix, Content: messages}
}

func (logs *Logs) Info(messages ...interface{})  {
	logs.messagesChannel <- Message{Prefix: infoPrefix, Content: messages}
}

func (logs *Logs) Warning(messages ...interface{})  {
	logs.messagesChannel <- Message{Prefix: warningPrefix, Content: messages}
}

func New(config Config) (*Logs, error) {
	var logs Logs
	logs.config = config
	logs.messagesChannel = make(chan Message)

	err := logs.setupLogger()
	if err != nil {
		return nil, err
	}

	logfileSize, err := getFileSize(logs.currentLogfile)
	if err != nil {
		return nil, err
	}
	logs.currentLogFileSize = logfileSize

	go logs.runWorker()

	return &logs, nil
}

func (logs *Logs) setupLogger() error {
	logfile, err := logs.openLogfile(logs.config.LogsDirPath)
	if err != nil {
		return err
	}
	logger := createLogger(logfile, "")

	logs.logger = logger
	logs.currentLogfile = logfile
	return nil
}

func (logs *Logs) openLogfile(directoryPath string) (*os.File, error) {
	filePath := logs.logfilePath(logfileBaseName)
	return os.OpenFile(filePath, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0644)
}

func (logs *Logs) logfilePath(fileName string) string {
	return filePath(logs.config.LogsDirPath, fileName)
}

func filePath(dirPath, fileName string) string {
	normalizedDirPath := strings.TrimRight(dirPath, "/")
	return strings.Join([]string{normalizedDirPath, fileName}, "/")
}

func createLogger(writer io.Writer, prefix string) *log.Logger {
	return log.New(writer, prefix, log.Ldate|log.Ltime)
}

func getFileSize(file *os.File) (int64, error) {
	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}
	return stat.Size(), err
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
	logs.logger.SetPrefix(message.Prefix)
	logs.logger.Println(message.Content...)

	logs.currentLogFileSize = logs.updatedLogFileSize(message.Content...)
	logs.changeLogfileIfTooBig(logs.currentLogFileSize)
}

func (logs *Logs) updatedLogFileSize(messages ...interface{}) int64 {
	//TODO calculate summary length of messages converted to string
	return logs.currentLogFileSize + int64(len(messages))
}

func (logs *Logs) changeLogfileIfTooBig(fileSize int64) error {
	if fileSize < logs.config.LogfileSplitThreshold {
		return nil
	}

	archivalLogFileName, err := logs.getNextLogfileName()
	if err != nil {
		return err
	}

	pathToArchivalLogFile := filePath(logs.config.LogsDirPath, archivalLogFileName)
	err = os.Rename(logs.logfilePath(logfileBaseName), pathToArchivalLogFile)
	if err != nil {
		return err
	}

	logs.currentLogfile.Close()
	logs.currentLogFileSize = 0
	if err = logs.setupLogger(); err != nil {
		return err
	}

	go logs.sendToS3(archivalLogFileName)

	return nil
}

func (logs *Logs) getNextLogfileName() (string, error) {
	logfileNumber, err := logs.getNextLogNumber(logs.config.LogsDirPath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.%d", logfileBaseName, logfileNumber), nil
}

func (logs *Logs) getNextLogNumber(directoryPath string) (int, error) {
	files, err := ioutil.ReadDir(logs.config.LogsDirPath)
	if err != nil {
		return 0, err
	}
    regex := regexp.MustCompile(logfileBaseName + "\\.([0-9]+)")

    highestLogNumber := 0
	for _, file := range files {
		match := regex.FindStringSubmatch(file.Name())
		logfileNumber, err := strconv.Atoi(match[1])
		if err != nil && highestLogNumber < logfileNumber {
			highestLogNumber = logfileNumber
		}
	}
	return highestLogNumber + 1, nil
}

func (logs *Logs) sendToS3(pathToFile string) {
	file, err := os.OpenFile(pathToFile, os.O_RDONLY, 0400)
	if err != nil {
		return
	}
	defer file.Close()

	// TODO call s3 api here
}
