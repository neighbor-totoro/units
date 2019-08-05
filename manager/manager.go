package manager

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"net"
	"time"

	"github.com/nnsgmsone/protocol"
	"github.com/nnsgmsone/units/breaker"
	"github.com/nnsgmsone/units/manager/tenant"
)

func init() {
	dealRegister = make(map[string]dealFunc)

	dealRegister["delRoom"] = deal0
	dealRegister["addRoom"] = deal1
	dealRegister["chgRoom"] = deal2

	dealRegister["rec"] = deal3
	dealRegister["rent"] = deal4

	dealRegister["list"] = deal5
}

func New(port int, ten tenant.Tenant, brk breaker.Breaker, timeout time.Duration) *manager {
	mg := new(manager)
	srv := protocol.New(port, mg, brk, dealMessage)
	mg.ten = ten
	mg.srv = srv
	mg.tm = timeout
	return mg
}

func (m *manager) Run() {
	m.srv.Run()
}

func (m *manager) Stop() {
	m.srv.Stop()
}

func dealMessage(usr interface{}, mw protocol.MessageWriter, msg *protocol.Message) {
	mg := usr.(*manager)
	switch m := msg.Msg.(type) {
	case protocol.MessageArray:
		f, ok := dealRegister[m.M[0]]
		if !ok {
			log.Printf("Illegal Message: %v\n", m.M)
			return
		}
		f(mg, msg.Name, mw, m.M[1:])
	default:
		log.Printf("Illegal Message Type: %T\n", m)
		return
	}
}

// delRoom name
func deal0(mg *manager, name string, mw protocol.MessageWriter, args []string) {
	if len(args) < 1 {
		mw.Write(name, errors.New("wrong number of arguments for 'delRoom'"))
		return
	}
	rs, err := mg.ten.DelRoom(args[0])
	if err != nil {
		mw.Write(name, err)
		return
	}
	for _, v := range rs {
		if conn, err := net.Dial("tcp", v); err == nil {
			conn.SetWriteDeadline(time.Now().Add(mg.tm))
			protocol.NewMessageWriter(bufio.NewWriter(conn)).Write("", protocol.DelRoom(args[0]))
			conn.Close()
		}
	}
	mw.Write(name, int64(0))
}

// addRoom name number
func deal1(mg *manager, name string, mw protocol.MessageWriter, args []string) {
	if len(args) < 2 {
		mw.Write(name, errors.New("wrong number of arguments for 'addRoom'"))
		return
	}
	mg.ten.AddRoom(args[0], args[1])
	mw.Write(name, int64(0))
}

// chgRoom name number
func deal2(mg *manager, name string, mw protocol.MessageWriter, args []string) {
	if len(args) < 2 {
		mw.Write(name, errors.New("wrong number of arguments for 'chgRoom'"))
		return
	}
	rs, err := mg.ten.DelRoom(args[0])
	if err != nil {
		mw.Write(name, err)
		return
	}
	for _, v := range rs {
		if conn, err := net.Dial("tcp", v); err == nil {
			conn.SetWriteDeadline(time.Now().Add(mg.tm))
			protocol.NewMessageWriter(bufio.NewWriter(conn)).Write("", protocol.DelRoom(args[0]))
			conn.Close()
		}
	}
	mg.ten.AddRoom(args[0], args[1])
	mw.Write(name, int64(0))
}

// recycle name renter
func deal3(mg *manager, name string, mw protocol.MessageWriter, args []string) {
	if len(args) < 2 {
		mw.Write(name, errors.New("wrong number of arguments for 'recycle'"))
		return
	}
	mg.ten.Recycle(args[0], args[1])
	mw.Write(name, int64(0))
}

// rent name renter
func deal4(mg *manager, name string, mw protocol.MessageWriter, args []string) {
	if len(args) < 2 {
		mw.Write(name, errors.New("wrong number of arguments for 'rent'"))
		return
	}
	num, err := mg.ten.Rent(args[0], args[1])
	if err != nil {
		mw.Write(name, err)
		return
	}
	mw.Write(name, num)
}

// list
func deal5(mg *manager, name string, mw protocol.MessageWriter, args []string) {
	rs, err := mg.ten.Rooms()
	if err != nil {
		mw.Write(name, err)
		return
	}
	ns := []string{}
	rts := [][]string{}
	for _, v := range rs {
		if n, err := mg.ten.RoomNumber(v); err != nil {
			mw.Write(name, err)
			return
		} else {
			ns = append(ns, n)
		}
		if rt, err := mg.ten.Renters(v); err != nil {
			mw.Write(name, err)
			return
		} else {
			rts = append(rts, rt)
		}
	}
	rl := new(RoomList)
	for i, j := 0, len(rs); i < j; i++ {
		rl.Rooms = append(rl.Rooms, &Room{
			Name:    rs[i],
			Number:  ns[i],
			Renters: rts[i],
		})
	}
	if data, err := json.Marshal(rl); err != nil {
		mw.Write(name, err)
	} else {
		mw.Write(name, string(data))
	}
}
