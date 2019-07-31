package unit

import (
	"net"
	"time"

	"github.com/nnsgmsone/protocol"
)

type Unit interface {
	Close()
	Send(string, interface{}) error
	SendAndRecv(string, interface{}) (*protocol.Message, error)
}

type unit struct {
	addr    string
	conn    net.Conn
	timeout time.Duration
}
