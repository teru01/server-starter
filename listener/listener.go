package listener

import (
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/hcl/strconv"
)

const ServerStarterEnvVarName = "SERVER_STARTER_PORT"

var (
	ErrNoListerningTarget = errors.New("No listening targer")
)

type Listener interface {
	Fd() uintptr
	Listen() (net.Listener, error)
	String() string
}

type ListenerList []Listener

func (ll ListenerList) String() string {
	list := make([]string, len(ll))
	for i, l := range ll {
		list[i] = l.String()
	}
	return strings.Join(list, ";")
}

type TCPListener struct {
	Addr string
	Port int
	fd uintptr
}

type UnixListener struct {
	Path string
	fd uintptr
}

func (l TCPListener) String() string {
	if l.Addr == "0.0.0.0" {
		return fmt.Sprintf("%d=%d", l.Port, l.fd)
	}
	return fmt.Sprintf("%s:%d=%d", l.Addr, l.Port, l.fd)
}

func (l TCPListener) Listen() (net.Listener, error) {
	return net.FileListener(os.NewFile(l.Fd(), fmt.Sprintf("%s:%d", l.Addr, l.Port)))
}

func (l TCPListener) Fd() uintptr {
	return l.fd
}

func (l UnixListener) String() string {
	return fmt.Sprintf("%s=%d", l.Path, l.fd)
}

func (l UnixListener) Listen() (net.Listener, error){
	return net.FileListener(os.NewFile(l.Fd(), l.Path))
}

func (l UnixListener) Fd() uintptr {
	return l.fd
}

var reLooksLikeHostPort = regexp.MustCompile(`^(\d+):(\d+)$`)
var reLooksLikePort = regexp.MustCompile(`^\d+$`)

func parseListenTargets(str string) ([]Listener, error) {
	if str == "" {
		return nil, ErrNoListerningTarget
	}
	rawspec := strings.Split(str, ";")
	ret := make([]Listener, len(rawspec))
	for i, pairString := range rawspec {
		pair := strings.Split(pairString, "=")
		hostPort := strings.TrimSpace(pair[0])
		fdString := strings.TrimSpace(pair[1])
		fd, err := strconv.ParseUint(fdString, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to parse\n")
		}

		if matches := reLooksLikeHostPort.FindAllString(hostPort, -1); matches != nil {
			port, err := strconv.ParseInt(matches[1], 10, 0)
			if err != nil {
				return nil, err
			}
			ret[i] = TCPListener{
				Addr: matches[0],
				Port: int(port),
				fd: uintptr(fd),
			}
		} else if match := reLooksLikePort.FindString(hostPort); match != "" {
			port, err := strconv.ParseInt(match, 10, 0)
			if err != nil {
				return nil, err
			}
			ret[i] = TCPListener{
				Addr: "0.0.0.0",
				Port: int(port),
				fd: uintptr(fd),
			}
		} else {
			ret[i] = UnixListener{
				Path: hostPort,
				fd: uintptr(fd)
			}
		}
	}
	return ret, nil
}

func GetPortsSpecification() string {
	return os.Getenv(ServerStarterEnvVarName)
}

func Ports() ([]Listener, error) {
	return parseListenTargets(GetPortsSpecification())
}

func ListenAll() ([]net.Listener, error) {
	targets, err := parseListenTargets(GetPortsSpecification())
	if err != nil {
		return nil, err
	}

	ret := make([]net.Listener, len(targets))
	for i, target := range targets {
		ret[i], err = target.Listen()
		if err != nil {
			for x := 0; x < i; x++ {
				ret[x].Close()
			}
			return nil, err
		}
	}
	return ret, nil
}

