// +build !windows

package starter

import (
	"os"
	"syscall"
)

func init() {

}

func addPlatformDependentNiceSignalNames(v map[syscall.Signal]string) map[syscall.Signal]string {

}

func findWorker(pid int) *os.Process {

}
