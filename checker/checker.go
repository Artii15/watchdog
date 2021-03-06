package checker

import (
	"os/exec"
	"fmt"
	"time"
	"sync"
	"github.com/Artii15/watchdog/loggers"
	"github.com/Artii15/watchdog/aws/sns"
)

type checker struct {
	snsNotifier *sns.Notifier
	loggersObject *loggers.Logs
	servicesDeadDuringPrevChecks map[string]bool
	responseChannel chan<- bool
}

func New(snsNotifier *sns.Notifier, loggersObject *loggers.Logs, responseChannel chan<- bool) *checker {
	return &checker{
		snsNotifier:snsNotifier,
		loggersObject: loggersObject,
		servicesDeadDuringPrevChecks: make(map[string]bool),
		responseChannel: responseChannel,
	}
}

func (checker *checker) Check(config *Config) {
	newDeadServices := checker.getDeadServices(config)
	checker.handleNewDeadServices(newDeadServices, config)
	checker.responseChannel<- true
}

func (checker *checker) getDeadServices(config *Config) []string {
	var newDeadServices []string
	for _, serviceName := range config.ListOfServices {
		if isServiceRunning(serviceName) {
			checker.servicesDeadDuringPrevChecks[serviceName] = false
		} else {
			if checker.isNewDeadService(serviceName) {
				newDeadServices = append(newDeadServices, serviceName)
			}
			checker.servicesDeadDuringPrevChecks[serviceName] = true
		}
	}
	return newDeadServices
}

func (checker *checker) isNewDeadService(serviceName string) bool {
	wasDead, wasChecked := checker.servicesDeadDuringPrevChecks[serviceName]
	return !wasChecked || !wasDead
}

func (checker *checker) handleNewDeadServices(deadServices []string, config *Config) {
	waitGroup := &sync.WaitGroup{}
	for _, serviceName := range deadServices {
		waitGroup.Add(1)
		go func() {
			checker.logServiceFailure(serviceName)
			checker.tryToRecoverService(serviceName, config)
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
}

func (checker *checker) logServiceFailure(serviceName string)  {
	message := fmt.Sprintf("Service %s is now inactive", serviceName)
	checker.loggersObject.Warning(message)
	checker.snsNotifier.Notify(message)
}

func (checker *checker) tryToRecoverService(serviceName string, config *Config)  {
	isServiceActive := false
	attemptsDone := 0
	for ; attemptsDone < config.NumOfAttempts && !isServiceActive; attemptsDone++ {
		currentAttemptNo := attemptsDone + 1
		checker.loggersObject.Info("Attempting to restart", serviceName, "Attempt", currentAttemptNo)
		isServiceActive = restartAndCheckIfRunning(serviceName)
		if isServiceActive {
			successMessage := fmt.Sprintf("Service %s successfully restarted after %d attempts", serviceName, currentAttemptNo)
			checker.loggersObject.Info(successMessage)
			checker.snsNotifier.Notify(successMessage)
		} else {
			failureMessage := fmt.Sprintf("Service %s still not active after %d restarts", serviceName, currentAttemptNo)
			checker.loggersObject.Warning(failureMessage)
			if currentAttemptNo == config.NumOfAttempts {
				checker.snsNotifier.Notify(failureMessage)
			} else {
				<-time.NewTimer(time.Duration(config.NumOfSecWait) * time.Second).C
			}
		}
	}
}

func restartAndCheckIfRunning(serviceName string) bool {
	restart(serviceName)
	return isServiceRunning(serviceName)
}

func restart(serviceName string)  {
	runCommand("service", serviceName, "restart")
}

func isServiceRunning(serviceName string) bool {
	command := runCommand("service", serviceName, "status")
	return command.ProcessState.Success()
}

func runCommand(commandName string, args ...string) *exec.Cmd {
	command := exec.Command(commandName, args...)
	command.Run()
	return command
}
