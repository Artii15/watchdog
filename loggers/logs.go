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
	"github.com/Artii15/watchdog/aws/s3"
)

const (
	errorPrefix   = "ERROR: "
	warningPrefix = "WARNING: "
	infoPrefix    = "INFO: "
)

const logfileBaseName = "watchdog.log"


type Logs struct {
	config          Config
	messagesChannel chan Message
	logger          *log.Logger
	currentLogfile  *os.File
	uploader        *s3.Uploader
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
	logs.uploader = nil

	go logs.runWorker()

	return &logs, nil
}

func (logs *Logs) SetUploader(uploader *s3.Uploader)  {
	logs.uploader = uploader
}

func (logs *Logs) setupLogger() error {
	os.MkdirAll(logs.config.LogsDirPath, os.ModePerm)
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

	logs.changeLogfileIfTooBig(logs.logfileSize())
}

func (logs *Logs) logfileSize() int64 {
	logfileStat, err := logs.currentLogfile.Stat()
	if err != nil {
		return 0
	}
	return logfileStat.Size()
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
	if err = logs.setupLogger(); err != nil {
		return err
	}

	go logs.sendToS3(pathToArchivalLogFile)

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
		logfileNumber := readLogfileNumberFromMatch(match)
		if highestLogNumber <= logfileNumber {
			highestLogNumber = logfileNumber
		}
	}
	return highestLogNumber + 1, nil
}

func readLogfileNumberFromMatch(match []string) int {
	if len(match) == 0 {
		return 0
	}
	logfileNumber, err := strconv.Atoi(match[1])
	if err != nil {
		return 0
	}
	return logfileNumber
}

func (logs *Logs) sendToS3(pathToFile string) {
	file, err := os.OpenFile(pathToFile, os.O_RDONLY, 0400)
	if err != nil || logs.uploader == nil {
		return
	}
	logs.uploader.Upload(file)

	defer file.Close()

}
