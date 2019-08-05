package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/infinivision/store/bg"
	"github.com/nnsgmsone/units/breaker"
	"github.com/nnsgmsone/units/manager"
	"github.com/nnsgmsone/units/manager/tenant/local"
)

func main() {
	enlargelimit()
	cfg := breaker.DefaultConfig()
	cfg.MaxRequests = 100000 // 10w/10s
	mg := manager.New(8888, local.New(bg.New("tenant.db")), breaker.New(cfg), 2*time.Second)
	go func() {
		for {
			ch := make(chan os.Signal)
			signal.Notify(ch)
			sig := <-ch
			if sig.String() == "quit" || sig.String() == "killed" || sig.String() == "interrupt" {
				mg.Stop()
				os.Exit(0)
			}
		}
	}()
	mg.Run()
}

func enlargelimit() error {
	var rlimit syscall.Rlimit

	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit); err != nil {
		return err
	} else {
		rlimit.Cur = rlimit.Max
		return syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlimit)
	}
	return nil
}
