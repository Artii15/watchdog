package checker

import (
	"os/exec"
	"watchdog/aws/sns"
	"watchdog/loggers"
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
