package kvstore

import (
	"fmt"
	"time"

	"github.com/infinivision/store"
	"github.com/nnsgmsone/protocol"
	"github.com/nnsgmsone/units/relay"
)

func init() {
	dealRegister = make(map[string]dealFunc)

	dealRegister["del"] = deal0
	dealRegister["set"] = deal1
	dealRegister["get"] = deal2
}

func New(db store.Store, ry relay.Relay, name string) *kvstore {
	ch := make(chan struct{})
	if mch, err := ry.Join(name, 1024, 4*time.Second); err != nil {
		return nil
	} else {
		return &kvstore{db, ry, ch, mch}
	}
}

func (kv *kvstore) Run() {
	go kv.ry.Run()
	kv.dealMessage()
}

func (kv *kvstore) Stop() {
	kv.ry.Stop()
	kv.ch <- struct{}{}
	<-kv.ch
	close(kv.ch)
}

func (kv *kvstore) dealMessage() {
	for {
		select {
		case <-kv.ch:
			kv.ch <- struct{}{}
			return
		case msg := <-kv.mch:
			switch m := msg.M.Msg.(type) {
			case protocol.MessageArray:
				f, ok := dealRegister[m.M[0]]
				if !ok {
					msg.Ch <- protocol.NewMessage(msg.M.Name, fmt.Errorf("Illegal Message"))
					continue
				}
				f(kv, msg, m.M[1:])
			default:
				msg.Ch <- protocol.NewMessage(msg.M.Name, fmt.Errorf("Illegal Message Type"))
			}
		}
	}
}

// del k
func deal0(kv *kvstore, msg *relay.Message, args []string) {
	if len(args) < 1 {
		msg.Ch <- protocol.NewMessage(msg.M.Name, fmt.Errorf("Wrong Number of Arguments"))
		return
	}
	if err := kv.db.Del([]byte(args[0])); err != nil {
		msg.Ch <- protocol.NewMessage(msg.M.Name, err)
		return
	}
	msg.Ch <- protocol.NewMessage(msg.M.Name, int64(0))
}

// set k v
func deal1(kv *kvstore, msg *relay.Message, args []string) {
	if len(args) < 2 {
		msg.Ch <- protocol.NewMessage(msg.M.Name, fmt.Errorf("Wrong Number of Arguments"))
		return
	}
	if err := kv.db.Set([]byte(args[0]), []byte(args[1])); err != nil {
		msg.Ch <- protocol.NewMessage(msg.M.Name, err)
		return
	}
	msg.Ch <- protocol.NewMessage(msg.M.Name, int64(0))
}

// get k
func deal2(kv *kvstore, msg *relay.Message, args []string) {
	if len(args) < 1 {
		msg.Ch <- protocol.NewMessage(msg.M.Name, fmt.Errorf("Wrong Number of Arguments"))
		return
	}
	if v, err := kv.db.Get([]byte(args[0])); err != nil {
		msg.Ch <- protocol.NewMessage(msg.M.Name, err)
	} else {
		msg.Ch <- protocol.NewMessage(msg.M.Name, string(v))
	}
}
