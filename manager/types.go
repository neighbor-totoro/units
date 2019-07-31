package manager

import (
	"time"

	"github.com/nnsgmsone/protocol"
	"github.com/nnsgmsone/units/manager/tenant"
)

type Manager interface {
	Run()
	Stop()
}

type manager struct {
	tm  time.Duration
	ten tenant.Tenant
	srv protocol.Server
}

type dealFunc (func(*manager, string, protocol.MessageWriter, []string))

var dealRegister map[string]dealFunc
