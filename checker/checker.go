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
}

func New(snsNotifier *sns.Notifier, loggersObject *loggers.Loggers) *checker {
	return &checker{isLastCheckingFinished: true, snsNotifier:snsNotifier, loggersObject: loggersObject}
}

func (checker *checker) Check(config *Config) {
	if checker.isLastCheckingFinished {
		checker.isLastCheckingFinished = false
		//deadServices := checker.getDeadServices(config)
		checker.isLastCheckingFinished = true
	}
}

func (checker *checker) getDeadServices(config *Config) []string {
	deadServices := make([]string, len(config.ListOfServices))
	for _, serviceName := range config.ListOfServices {
		if !isServiceRunning(serviceName) {
			deadServices = append(deadServices, serviceName)
		}
	}
	return deadServices
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
