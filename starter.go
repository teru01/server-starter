package starter

import (
	"net"
	"os"
	"syscall"
	"time"
)

var niceSigNames map[syscall.Signal]string
var niceNameToSigs map[string]syscall.Signal
var successStatus syscall.WaitStatus
var failureStatus syscall.WaitStatus

func makeNiceSigNamesCommon() map[syscall.Signal]string {

}

func makeNiceSigNames() map[syscall.Signal]string {

}

func init() {

}

type listener struct {
	listener net.Listener
	spec string
}

type Config interface {
	Args() []string
	Command() string
	Dir() string
	Interval() time.Duration
	PidFile() string
	Ports() []string
	Paths() []string
	SignalOnHUP() os.Signal
	SignalOnTERM() os.Signal
	StatusFile() string
}

type Starter struct {
	interval time.Duration
	signalOnHUP os.Signal
	signalOnTERM os.Signal
	statusFile string
	pidFile    string
	dir        string
	ports      []string
	paths      []string
	listeners  []listener
	generation int
	command    string
	args       []string
}

func NewStarter(c Config) (*Starter, error) {

}

func (s Starter) Stop() {

}

type processState interface {
	Pid() int
	Sys() interface{}
}

type dummyProcessState struct {
	pid int
	status syscall.WaitStatus
}

func grabExitStatus(st processState) syscall.WaitStatus {

}

func (d dummyProcessState) Pid() int {
	return d.Pid()
}

func (d dummyProcessState) Sys() interface{} {
	return d.status
}

func signame(s os.Signal) string {

}

func SigFromName(n string) os.Signal {

}

func setEnv() {

}

func parsePortSpec(addr string) (string, int, error) {

}

func (s *Starter) Run() error {

}

func getKillOldDelay() time.Duration {

}

type WorkerState int

const (
	WorkerStarted WorkerState = iota
	ErrFailedToStart
)

func (s *Starter) StartWorker(sigCh chan os.Signal, ch chan processState) *os.Process {

}

func (s *Starter) Teardown() error {
	
}
