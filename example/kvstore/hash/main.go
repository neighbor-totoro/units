package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/nnsgmsone/units/breaker"
	"github.com/nnsgmsone/units/example/kvstore/hash/kvhash"
	"github.com/nnsgmsone/units/relay"
)

func main() {
	enlargelimit()
	cfg := breaker.DefaultConfig()
	cfg.MaxRequests = 100000 // 10w/10s
	ry := relay.New(&relay.Config{8888, "192.168.0.5:8888", []string{"192.168.0.2:8888"}, 2 * time.Second}, breaker.New(cfg), nil)
	kh := kvhash.New(ry, "hash", []string{"store1", "store2"})
	go func() {
		for {
			ch := make(chan os.Signal)
			signal.Notify(ch)
			sig := <-ch
			if sig.String() == "quit" || sig.String() == "killed" || sig.String() == "interrupt" {
				kh.Stop()
				os.Exit(0)
			}
		}
	}()
	kh.Run()
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
