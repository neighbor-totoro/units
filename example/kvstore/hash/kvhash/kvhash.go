package kvhash

import (
	"fmt"
	"hash/crc32"
	"time"

	"github.com/nnsgmsone/protocol"
	"github.com/nnsgmsone/units/relay"
)

func init() {
	dealRegister = make(map[string]dealFunc)

	dealRegister["del"] = deal0
	dealRegister["set"] = deal1
	dealRegister["get"] = deal2
}

func New(ry relay.Relay, name string, rg []string) *kvhash {
	ch := make(chan struct{})
	if mch, err := ry.Join(name, 1024, 4*time.Second); err != nil {
		return nil
	} else {
		return &kvhash{rg, ry, ch, mch}
	}
}

func (kv *kvhash) Run() {
	go kv.ry.Run()
	kv.dealMessage()
}

func (kv *kvhash) Stop() {
	kv.ry.Stop()
	kv.ch <- struct{}{}
	<-kv.ch
	close(kv.ch)
}

func (kv *kvhash) dealMessage() {
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
				f(kv, msg, m.M)
			default:
				msg.Ch <- protocol.NewMessage(msg.M.Name, fmt.Errorf("Illegal Message Type"))
			}
		}
	}
}

// del k
func deal0(kv *kvhash, msg *relay.Message, args []string) {
	if len(args) < 2 {
		msg.Ch <- protocol.NewMessage(msg.M.Name, fmt.Errorf("Wrong Number of Arguments"))
		return
	}
	h := crc32.NewIEEE()
	h.Write([]byte(args[1]))
	msg.M.Name = kv.rg[int(h.Sum32())%len(kv.rg)]
	kv.ry.Channel() <- msg
}

// set k v
func deal1(kv *kvhash, msg *relay.Message, args []string) {
	if len(args) < 3 {
		msg.Ch <- protocol.NewMessage(msg.M.Name, fmt.Errorf("Wrong Number of Arguments"))
		return
	}
	h := crc32.NewIEEE()
	h.Write([]byte(args[1]))
	msg.M.Name = kv.rg[int(h.Sum32())%len(kv.rg)]
	kv.ry.Channel() <- msg
}

// get k
func deal2(kv *kvhash, msg *relay.Message, args []string) {
	if len(args) < 2 {
		msg.Ch <- protocol.NewMessage(msg.M.Name, fmt.Errorf("Wrong Number of Arguments"))
		return
	}
	h := crc32.NewIEEE()
	h.Write([]byte(args[1]))
	msg.M.Name = kv.rg[int(h.Sum32())%len(kv.rg)]
	kv.ry.Channel() <- msg
}
