package relay

import (
	"sync"
	"time"

	"github.com/nnsgmsone/protocol"
	"github.com/nnsgmsone/units/unit"
)

type Relay interface {
	Run()
	Stop()

	Exit(string) error
	Join(string, int, time.Duration) (<-chan *Message, error)

	Channel() chan<- *Message // 如果设置了Ch则会得到一个回复
}

type Callback (func(string)) // unit被注销时会调用该回调函数

type Message struct {
	M  *protocol.Message
	Ch chan *protocol.Message
}

type Config struct {
	Port     int
	Address  string
	Managers []string
	Timeout  time.Duration
}

type relayManager struct {
	c       uint     // current manager
	mgs     []string // manager server list
	u       unit.Unit
	timeout time.Duration
}

type relayUnits struct {
	us      *sync.Map // name -> unit
	cs      *sync.Map // name -> unit's channel
	timeout time.Duration
}

type relay struct {
	relayUnits
	relayManager
	addr     string
	callback Callback
	ch       chan struct{}
	mch      chan *Message
	srv      protocol.Server
}

type unitChannel struct {
	ch      chan *Message
	timeout time.Duration
}

type dealFunc (func(*relay, []string))

var dealRegister map[string]dealFunc
