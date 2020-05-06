package starter

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
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
	spec     string
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
	interval     time.Duration
	signalOnHUP  os.Signal
	signalOnTERM os.Signal
	statusFile   string
	pidFile      string
	dir          string
	ports        []string
	paths        []string
	listeners    []listener
	generation   int
	command      string
	args         []string
}

func NewStarter(c Config) (*Starter, error) {
	if c == nil {
		return nil, fmt.Errorf("config argument must be non-nil")
	}
	var signalOnHUP os.Signal = syscall.SIGTERM
	var signalOnTERM os.Signal = syscall.SIGTERM
	if s := c.SignalOnHUP(); s != nil {
		signalOnHUP = s
	}
	if s := c.SignalOnTERM(); s != nil {
		signalOnTERM = s
	}
	if c.Command() == "" {
		return nil, fmt.Errorf("argument command must be specified.")
	}
	if _, err := exec.LookPath(c.Command()); err != nil {
		return nil, err
	}

	s := &Starter{
		args:         c.Args(),
		command:      c.Command(),
		dir:          c.Dir(),
		interval:     c.Interval(),
		listeners:    make([]listener, 0, len(c.Ports())+len(c.Paths())),
		pidFile:      c.PidFile(),
		ports:        c.Ports(),
		paths:        c.Paths(),
		signalOnHUP:  signalOnHUP,
		signalOnTERM: signalOnTERM,
		statusFile:   c.StatusFile(),
	}

	return s, nil
}

func (s Starter) Stop() {
	p, _ := os.FindProcess(os.Getpid())
	p.Signal(syscall.SIGTERM)
}

type processState interface {
	Pid() int
	Sys() interface{}
}

type dummyProcessState struct {
	pid    int
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
	defer s.Teardown()

	if s.pidFile != nil {
		f, err := os.OpenFile(s.pidFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
			return err
		}
		fmt.Fprintf(f, "%d", os.Getegid())
		defer func() {
			os.Remove(f.Name())
			f.Close()
		}()
	}

	for _, addr := range s.ports {
		var l net.Listener
		host, port, err := parsePortSpec(addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse addr spec '%s': %s\n", addr, err)
			return err
		}

		hostport := fmt.Sprintf("%s:%d", host, port)
		l, err = net.Listen("tcp4", hostport)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to listen on %s:%s\n", hostport, err)
			return err
		}

		spec := ""
		if host == "" {
			spec = fmt.Sprint("%d", port)
		} else {
			spec = fmt.Sprintf("%s:%d", host, port)
		}
		s.listeners = append(s.listeners, listener{listener: l, spec: spec})
	}

	// for unix domain sockets
	for _, path := range s.paths {
		var l net.Listener
		if fl, err := os.Lstat(path); err == nil && fl.Mode()&os.ModeSocket == os.ModeSocket { // A&B==Bの時，BがONならAもON，BがOFFの時はAは任意
			fmt.Fprintf(os.Stderr, "removing existing socket file:%s\n", path)
			err = os.Remove(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to remove existing socket file:%s:%s\n", path, err)
				return err
			}
		}
		_ = os.Remove(path) // 既存のソケットじゃないファイルを消してしまう？
		l, err := net.Listen("unix", path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to listen file %s:%s\n", path, err)
			return err
		}
		s.listeners = append(s.listeners, listener{listener: l, spec: path})
	}

	s.generation = 0
	os.Setenv("SERVER_STARTER_GENERATION", fmt.Sprintf("%d", s.generation))

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	setEnv()
	workerCh := make(chan processState)
	p := s.StartWorker(sigCh, workerCh)
	oldWorkers := make(map[int]int)
	var sigReceived os.Signal
	var sigToSend os.Signal

	statusCh := make(chan map[int]int)

	// open file and save status
	go func(filename string, ch chan map[int]int) {
		for wmap := range ch {
			if filename == "" {
				continue
			}

			f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			if err != nil {
				continue
			}

			for gen, pid := range wmap {
				fmt.Fprintf(f, "%d:%d\n", gen, pid)
			}
			f.Close()
		}
	}(s.statusFile, statusCh)

	defer func() {
		if p != nil {
			oldWorkers[p.Pid] = s.generation
		}

		fmt.Fprintf(os.Stderr, "received %s, sending %s to all workers:", signame(sigReceived), signame(sigToSend))
		printMap(oldWorkers)

		for pid := range oldWorkers {
			worker, err := os.FindProcess(pid)
			if err != nil {
				continue
			}
			worker.Signal(sigToSend)
		}

		for len(oldWorkers) > 0 {
			st := <- workerCh
			fmt.Fprintf(os.Stderr, "worker %d died, status: %d\n", st.Pid(), grabExitStatus(st))
			delete(oldWorkers, st.Pid())
		}
		fmt.Fprintf(os.Stderr, "exiting\n")
	}()

	for {
		setEnv()
		for {
			status := make(map[int]int)
			for pid, gen := range oldWorkers {
				status[gen] = pid
			}
			status[s.generation] = p.Pid
			statusCh <- status

			restart := 0
			select {
			case st := <- workerCh:
				if p.Pid == st.Pid() {
					exitSt := grabExitStatus(st)
					fmt.Fprintf(os.Stderr, "worker %d died unexpectedly with status %d, restarting\n", p.Pid, exitSt)
					p = s.StartWorker(sigCh, workerCh)
				} else {
					exitSt := grabExitStatus(st)
					fmt.Fprintf(os.Stderr, "old worker %d died, status: %d\n", st.Pid(), exitSt)
					delete(oldWorkers, st.Pid())
				}
			case sigReceived = <- sigCh:
				switch sigReceived {
				case syscall.SIGHUP:
					fmt.Fprintf(os.Stderr, "received HUP (num_old_workers=TODO)\n")
					restart = 1
					sigToSend = s.signalOnHUP
				case syscall.SIGTERM:
					sigToSend = s.signalOnTERM
					return nil
				default:
					sigToSend = syscall.SIGTERM
					return nil
				}
			}

			if restart > 1 || restart > 0 && len(oldWorkers) == 0 {
				fmt.Fprintf(os.Stderr, "spawning a new worker (num_old_workers=TODO\n")
				oldWorkers[p.Pid] = s.generation
				p = s.StartWorker(sigCh, workerCh)
				fmt.Fprintf(os.Stderr, "new worker is now running, sending %s to old workers:", signame(sigToSend))
				size := len(oldWorkers)
				if size == 0 {
					fmt.Fprintf(os.Stderr, "none\n")
				} else {
					printMap(oldWorkers)

					killOldDelay := getKillOldDelay()
					fmt.Fprintf(os.Stderr, "sleep %d secs\n", int(killOldDelay/time.Second))
					if killOldDelay > 0 {
						time.Sleep(killOldDelay)
					}
					fmt.Fprintf(os.Stderr, "killing old workers\n")

					for pid := range oldWorkers {
						worker, err := os.FindProcess(pid)
						if err != nil {
							continue
						}
						worker.Signal(s.signalOnHUP)
					}
				}
			}
		}
	}
}

func printMap(m map[int]int) {
	for i, pid := range m {
		fmt.Fprintf(os.Stderr, "%d", pid)
		if i < len(m)-1 {
			fmt.Fprintf(os.Stderr, ",")
		}
	}
	fmt.Fprintf(os.Stderr, "\n")
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
