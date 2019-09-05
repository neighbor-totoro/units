package kvhash

import (
	"github.com/nnsgmsone/units/relay"
)

type kvhash struct {
	rg  []string // region
	ry  relay.Relay
	ch  chan struct{}
	mch <-chan *relay.Message
}

type dealFunc (func(*kvhash, *relay.Message, []string))

var dealRegister map[string]dealFunc
