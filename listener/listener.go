package listener

import (
	"errors"
	"net"
	"regexp"
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

}

func (l TCPListener) Listen() (net.Listener, error) {

}

func (l TCPListener) Fd() uintptr {

}

func (l UnixListener) String() string {

}

func (l UnixListener) Listen() (net.Listener, error){

}

func (l UnixListener) Fd() uintptr {

}

var reLooksLikeHostPort = regexp.MustCompile(`^(\d+):(\d+)$`)
var reLooksLikePort = regexp.MustCompile(`^\d+$`)

func parseListenTargets(str string) ([]Listener, error) {

}

func GetPortsSpecification() string {

}

func Ports() ([]Listener, error) {

}

func ListenAll() ([]net.Listener, error) {

}

