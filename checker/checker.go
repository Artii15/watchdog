package checker

import "os/exec"

type Checker struct {
	Config *Config
	
}

func New(config *Config) *Checker {
	return &Checker{Config:config}
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
