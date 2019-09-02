package relay

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nnsgmsone/protocol"
	"github.com/nnsgmsone/units/breaker"
	"github.com/nnsgmsone/units/manager"
	"github.com/nnsgmsone/units/unit"
)

func init() {
	dealRegister = make(map[string]dealFunc)

	dealRegister["delRoom"] = deal0
}

func New(cfg *Config, brk breaker.Breaker, callback Callback) *relay {
	r := new(relay)
	mg := newManager(cfg.Managers, cfg.Timeout)
	srv := protocol.New(cfg.Port, r, brk, dealMessage)
	us := relayUnits{new(sync.Map), new(sync.Map), cfg.Timeout}
	r.srv = srv
	r.relayUnits = us
	r.relayManager = mg
	r.addr = cfg.Address
	r.callback = callback
	r.ch = make(chan struct{})
	r.mch = make(chan *Message)
	return r
}

// manager name is empty
func (r *relay) Exit(name string) error {
	r.cs.Delete(name)
	return nil
}

// 重启后所有的本地节点都要重新调用join
func (r *relay) Join(name string, size int, timeout time.Duration) (<-chan *Message, error) {
	uc := &unitChannel{make(chan *Message, size), timeout}
	r.cs.Store(name, uc)
	return uc.ch, nil
}

func (r *relay) Channel() chan<- *Message {
	return r.mch
}

func (r *relay) Run() {
	go r.srv.Run()
	for {
		select {
		case <-r.ch:
			r.ch <- struct{}{}
			return
		case msg := <-r.mch:
			if v, ok := r.cs.Load(msg.M.Name); ok { // local unit
				v.(*unitChannel).ch <- msg
				continue
			}
			switch {
			case msg.Ch != nil:
				if m, err := r.sendAndRecv(msg.M.Name, msg.M); err != nil {
					msg.Ch <- protocol.NewMessage(msg.M.Name, err)
				} else {
					msg.Ch <- m
				}
			default:
				r.send(msg.M.Name, msg.M)
			}
		}
	}
}

func (r *relay) Stop() {
	r.srv.Stop()
	r.ch <- struct{}{}
	<-r.ch
	close(r.ch)
	close(r.mch)
}

func (r *relay) send(name string, msg *protocol.Message) error {
	u, ok := r.us.Load(name)
	if !ok {
		if err := r.relayManager.getUnit(name, r); err != nil {
			return fmt.Errorf("Cannot Rent %s: %v", name, err)
		}
		if u, ok = r.us.Load(name); !ok {
			return errors.New("Positioning Failed")
		}
	}
	var err error
	switch m := msg.Msg.(type) {
	case protocol.MessageError:
		err = u.(unit.Unit).Send(name, m.M)
	case protocol.MessageArray:
		err = u.(unit.Unit).Send(name, m.M)
	default:
		return errors.New("Unsupport TYpe")
	}
	if err != nil {
		r.relayManager.getUnit(name, r)
		return err
	}
	return nil
}

func (r *relay) sendAndRecv(name string, msg *protocol.Message) (*protocol.Message, error) {
	u, ok := r.us.Load(name)
	if !ok {
		if err := r.relayManager.getUnit(name, r); err != nil {
			return nil, fmt.Errorf("Cannot Rent %s: %v", name, err)
		}
		if u, ok = r.us.Load(name); !ok {
			return nil, errors.New("Positioning Failed")
		}
	}
	var err error
	var resp *protocol.Message
	switch m := msg.Msg.(type) {
	case protocol.MessageError:
		resp, err = u.(unit.Unit).SendAndRecv(name, m.M)
	case protocol.MessageArray:
		resp, err = u.(unit.Unit).SendAndRecv(name, m.M)
	default:
		return nil, errors.New("Unsupport TYpe")
	}
	if err != nil {
		r.relayManager.getUnit(name, r)
		return nil, err
	}
	return resp, nil
}

func dealMessage(usr interface{}, mw protocol.MessageWriter, msg *protocol.Message) {
	r := usr.(*relay)
	switch {
	case msg.Name == "":
		switch m := msg.Msg.(type) {
		case protocol.MessageArray:
			{
				fmt.Printf("Mesage: %v\n", m.M)
			}
			f, ok := dealRegister[m.M[0]]
			if !ok {
				log.Printf("Illegal Message: %v\n", m.M)
				return
			}
			f(r, m.M[1:])
		default:
			log.Printf("Illegal Message Type: %T\n", m)
		}
	default:
		v, ok := r.cs.Load(msg.Name)
		if !ok {
			return
		}
		uc := v.(*unitChannel)
		ch := make(chan *protocol.Message)
		uc.ch <- &Message{M: msg, Ch: ch}
		select {
		case m := <-ch:
			mw.WriteMessage(m)
		case <-time.After(uc.timeout):
		}
		close(ch)
	}
}

// delRoom name
func deal0(r *relay, args []string) {
	if len(args) < 1 {
		return
	}
	{
		fmt.Printf("delRoom %s\n", args[0])
	}
	r.us.Delete(args[0])
	r.cs.Delete(args[0])
	if r.callback != nil {
		r.callback(args[0])
	}
}

func newManager(mgs []string, timeout time.Duration) relayManager {
	return relayManager{0, mgs, unit.New(mgs[0], timeout), timeout}
}

func (rm *relayManager) retry() {
	rm.c++
	rm.u.Close()
	rm.u = unit.New(rm.mgs[rm.c%uint(len(rm.mgs))], rm.timeout)
}

func (rm *relayManager) getUnit(name string, r *relay) error {
	var err error
	var msg *protocol.Message

	if msg, err = rm.u.SendAndRecv("", manager.Rent(name, r.addr)); err != nil {
		rm.retry()
		msg, err = rm.u.SendAndRecv("", manager.Rent(name, r.addr))
	}
	if err != nil {
		return err
	}
	switch m := msg.Msg.(type) {
	case protocol.MessageArray:
		r.us.Store(name, unit.New(m.M[0], rm.timeout))
		return nil
	case protocol.MessageError:
		return errors.New(m.M)
	default:
		return errors.New("Unexpected Response")
	}
}
