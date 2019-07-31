package unit

import (
	"bufio"
	"net"
	"time"

	"github.com/nnsgmsone/protocol"
)

func New(addr string, timeout time.Duration) *unit {
	u := new(unit)
	u.addr = addr
	u.timeout = timeout
	u.conn, _ = net.Dial("tcp", u.addr)
	return u
}

func (u *unit) Send(name string, msg interface{}) error {
	if err := u.reConnect(); err != nil {
		return err
	}
	u.conn.SetWriteDeadline(time.Now().Add(u.timeout))
	switch err := protocol.NewMessageWriter(bufio.NewWriter(u.conn)).Write(name, msg); err {
	case nil:
		return nil
	case protocol.TYPEERROR, protocol.ENCODEERROR:
		return err
	default: // retry
		if err = u.connect(); err != nil {
			return err
		}
		u.conn.SetWriteDeadline(time.Now().Add(u.timeout))
		return protocol.NewMessageWriter(bufio.NewWriter(u.conn)).Write(name, msg)
	}
}

func (u *unit) SendAndRecv(name string, msg interface{}) (*protocol.Message, error) {
	if err := u.Send(name, msg); err != nil {
		return nil, err
	}
	return u.recv()
}

func (u *unit) Close() {
	if u.conn != nil {
		u.conn.Close()
		u.conn = nil
	}
}

func (u *unit) recv() (*protocol.Message, error) {
	u.conn.SetReadDeadline(time.Now().Add(u.timeout))
	return protocol.ReadMessage(bufio.NewReader(u.conn))
}

func (u *unit) connect() error {
	var err error

	if u.conn != nil {
		u.conn.Close()
		u.conn = nil
	}
	if u.conn, err = net.Dial("tcp", u.addr); err != nil {
		return err
	}
	return nil
}

func (u *unit) reConnect() error {
	if u.conn == nil {
		return u.connect()
	}
	return nil
}
