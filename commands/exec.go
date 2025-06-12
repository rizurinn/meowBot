package commands

import (
    "context"
    "fmt"
    "os/exec"
    "time"
)

func ExecuteCommand(command string, args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return string(output), fmt.Errorf("command failed: %v", err)
	}
	
	return string(output), nil
}
