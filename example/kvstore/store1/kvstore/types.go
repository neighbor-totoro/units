package kvstore

import (
	"github.com/infinivision/store"
	"github.com/nnsgmsone/units/relay"
)

type kvstore struct {
	db  store.Store
	ry  relay.Relay
	ch  chan struct{}
	mch <-chan *relay.Message
}

type dealFunc (func(*kvstore, *relay.Message, []string))

var dealRegister map[string]dealFunc
