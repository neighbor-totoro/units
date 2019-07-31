package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/infinivision/store/bg"
	"github.com/nnsgmsone/units/breaker"
	"github.com/nnsgmsone/units/relay"
	"github.com/nnsgmsone/units/unit/kvstore"
)

func main() {
	cfg := breaker.DefaultConfig()
	cfg.MaxRequests = 100000 // 10w/10s
	db := bg.New("store.db")
	ry := relay.New(&relay.Config{8888, "192.168.0.3:8888", []string{"192.168.0.2:8888"}, 2 * time.Second}, breaker.New(cfg), nil)
	ks := kvstore.New(db, ry, "store1")
	go func() {
		for {
			ch := make(chan os.Signal)
			signal.Notify(ch)
			sig := <-ch
			if sig.String() == "quit" || sig.String() == "killed" || sig.String() == "interrupt" {
				ks.Stop()
				os.Exit(0)
			}
		}
	}()
	ks.Run()
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
