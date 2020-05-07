package starter

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

var echoServerTxt = `package main

import (
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/teru01/server-starter/listener"
)

func main() {
	listeners, err := listener.ListenAll()
	if err != nil {
		panic(err)
	}
	handler := http.HandlerFunc(func (w http.RequestWriter, r *http.Request) {
		io.Copy(w, r.Body)
	})
	for _, l := range listeners {
		http.Serve(l, handler)
	}
	loop := false
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGHUP)
	for loop {
		select {
		case <- sigCh:
			loop = false
		default:
			time.Sleep(time.Second)
		}
	}
}`

type config struct {
	args       []string
	command    string
	dir        string
	interval   int
	pidfile    string
	ports      []string
	paths      []string
	sigonhup   string
	sigonterm  string
	statusfile string
}

func (c config) Args() []string          { return c.args }
func (c config) Command() string         { return c.command }
func (c config) Dir() string             { return c.dir }
func (c config) Interval() time.Duration { return time.Duration(c.interval) * time.Second }
func (c config) PidFile() string         { return c.pidfile }
func (c config) Ports() []string         { return c.ports }
func (c config) Paths() []string         { return c.paths }
func (c config) SignalOnHUP() os.Signal  { return SigFromName(c.sigonhup) }
func (c config) SignalOnTERM() os.Signal { return SigFromName(c.sigonterm) }
func (c config) StatusFile() string      { return c.statusfile }

func TestRun(t *testing.T) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("server-starter-test-%d", os.Getpid()))
	if err != nil {
		t.Errorf("failed to create temp %s", err)
		return
	}
	defer os.RemoveAll(dir)

	srcFile := filepath.Join(dir, "echod.go")
	f, err := os.OpenFile(srcFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		t.Errorf("faied to create %s: %s", srcFile, err)
		return
	}
	defer f.Close()
	io.WriteString(f, echoServerTxt)

	_, lastComp := filepath.Split(dir)
	fmt.Println("last: ", lastComp)
	cmd := exec.Command("go", "mod", "init", "github.com/teru01/server-starter/" + lastComp)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Logf("%s", output)
		t.Errorf("failed to run go mod init %s", err)
		return
	}

	executableName := filepath.Join(dir, "echod.go")
	cmd = exec.Command("go", "build", "-o", executableName, ".")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Logf("%s", output)
		t.Errorf("failed to compile %s", err)
		return
	}

	ports := []string{"9090", "8080"}
	sd, err := NewStarter(&config {
		ports: ports,
		command: executableName,
	})

	if err != nil {
		t.Errorf("failed to create starter %s", err)
		return
	}

	doneCh := make(chan struct{})
	readyCh := make(chan struct{})
	go func() {
		defer func () {
			doneCh <- struct{}{}
		}()
		time.AfterFunc(500*time.Millisecond, func() {
			readyCh <- struct{}{}
		})
		if err := sd.Run(); err != nil {
			t.Errorf("sd Run() failed %s", err)
		}
		t.Logf("exiting...")
	}()

	<-readyCh
	for _, port := range ports {
		_, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%s", port))
		if err != nil {
			t.Errorf("error connecting to port '%s:%s'", port, err)
		}
	}
	time.AfterFunc(time.Second, sd.Stop)
	<-doneCh

	patterns := make([]string, len(ports))
	for i, port := range ports {
		patterns[i] = fmt.Sprintf(`%s=\d+`, port)
	}
	pattern := regexp.MustCompile(strings.Join(patterns, ";"))

	if envPort := os.Getenv("SERVER_STARTER_PORT"); !pattern.MatchString(envPort) {
		t.Errorf("SERVER_STARTER_PORT expected %s, got %s", pattern, envPort)
	}
}

// func TestSigFromName(t *testing.T) {

// }
