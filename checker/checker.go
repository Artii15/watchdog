package checker

import (
	"os/exec"
	"watchdog/aws/sns"
	"watchdog/loggers"
	"fmt"
	"time"
	"sync"
)

type checker struct {
	isLastCheckingFinished bool
	snsNotifier *sns.Notifier
	loggersObject *loggers.Loggers
	servicesDeadDuringPrevChecks map[string]bool
}

func New(snsNotifier *sns.Notifier, loggersObject *loggers.Loggers) *checker {
	return &checker{
		isLastCheckingFinished: true,
		snsNotifier:snsNotifier,
		loggersObject: loggersObject,
		servicesDeadDuringPrevChecks: make(map[string]bool),
	}
}

func (checker *checker) Check(config *Config) {
	if checker.isLastCheckingFinished {
		checker.isLastCheckingFinished = false
		newDeadServices := checker.getDeadServices(config)
		checker.handleNewDeadServices(newDeadServices, config)
		checker.isLastCheckingFinished = true
	}
}

func (checker *checker) getDeadServices(config *Config) []string {
	newDeadServices := make([]string, len(config.ListOfServices))
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
	checker.loggersObject.Warning.Println(message)
	checker.snsNotifier.Notify(message)
}

func (checker *checker) tryToRecoverService(serviceName string, config *Config)  {
	attemptsTimer := time.NewTimer(time.Duration(config.NumOfSecWait) * time.Second)
	isServiceActive := false
	attemptsDone := 0
	for ; attemptsDone < config.NumOfAttempts && !isServiceActive; attemptsDone++ {
		currentAttemptNo := attemptsDone + 1
		checker.loggersObject.Info.Println("Attempting to restart", serviceName, "Attempt", currentAttemptNo)
		isServiceActive = restartAndCheckIfRunning(serviceName)
		if isServiceActive {
			checker.loggersObject.Info.Println("Service", serviceName, "successfully restarted after", currentAttemptNo, "attempts")
		} else {
			checker.loggersObject.Warning.Println("Service", serviceName, "still not active after", currentAttemptNo, "restarts")
			<-attemptsTimer.C
		}
	}
	attemptsTimer.Stop()
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
