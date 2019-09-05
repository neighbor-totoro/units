package manager

import (
	"time"

	"github.com/nnsgmsone/protocol"
	"github.com/nnsgmsone/units/manager/tenant"
)

type Room struct {
	Name    string   `json: "name"`
	Number  string   `json: "number"`
	Renters []string `json: renters`
}

type RoomList struct {
	Rooms []*Room
}

type manager struct {
	tm  time.Duration
	ten tenant.Tenant
	srv protocol.Server
}

type dealFunc (func(*manager, string, protocol.MessageWriter, []string))

var dealRegister map[string]dealFunc
