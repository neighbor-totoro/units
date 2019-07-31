package kventry

import "github.com/nnsgmsone/units/relay"

const (
	S0 = iota // add
	S1        // del
	S2        // get
)

type KvEntry interface {
	Run()
	Stop()
}

type HttpResult struct {
	Err string `json: "err"`
	Msg string `json: "msg"`
}

type kventry struct {
	port int    // http port
	hu   string // hash unit
	ry   relay.Relay
	ch   chan struct{}
	mch  <-chan *relay.Message
}
