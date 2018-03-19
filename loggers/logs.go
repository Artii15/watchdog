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

	logger, currentLogfile, err := setupLogger(config.logsDirPath)
	if err != nil {
		return nil, err
	}
	logs.logger = logger
	logs.currentLogfile = currentLogfile

	logfileSize, err := getFileSize(currentLogfile)
	if err != nil {
		return nil, err
	}
	logs.currentLogFileSize = logfileSize

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
	logs.logger.Println(message.Content)

	logs.currentLogFileSize = logs.updatedLogFileSize(message.Content)
	logs.changeLogfileIfTooBig(logs.currentLogFileSize)
}

func (logs *Logs) updatedLogFileSize(message string) int64 {
	return logs.currentLogFileSize + int64(len(message))
}

func (logs *Logs) changeLogfileIfTooBig(fileSize int64) {
	if fileSize > logs.config.logFileSplitThreshold {


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

func (logs *Logs) getNextLogfileName(directoryPath string) (string, error) {
	logfileNumber, err := logs.getNextLogNumber(directoryPath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.%d", logfileBaseName, logfileNumber), nil
}

func (logs *Logs) getNextLogNumber(directoryPath string) (int, error) {
	files, err := ioutil.ReadDir(logs.config.logsDirPath)
	if err != nil {
		return 0, err
	}
    regex := regexp.MustCompile(logfileBaseName + "\\.([0-9]+)")

    biggestLogNumber := 0
	for _, file := range files {
		match := regex.FindStringSubmatch(file.Name())
		logfileNumber, err := strconv.Atoi(match[1])
		if err != nil && biggestLogNumber < logfileNumber {
			biggestLogNumber = logfileNumber
		}
	}
	return biggestLogNumber+1, nil
}

func sendToS3(file *os.File) {
	//TODO
}
