package redeploy

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func Redeploy() error {
	cmd := exec.Command("./cplane/redeploy/redeploy.sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Start a new process group
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return err
	}

	fmt.Println("Running redeploy script...")

	return nil
}
