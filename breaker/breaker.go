package breaker

import (
	"errors"
	"sync/atomic"
	"time"
)

func New(cfg *Config) *breaker {
	ws := make([]window, cfg.WindowSize)
	return &breaker{
		Config: cfg,
		ws:     ws,
		s:      CLOSED,
		t:      time.Now().Unix(),
	}
}

func DefaultConfig() *Config {
	return &Config{
		WindowSize:     10,               // 10s
		MaxRequests:    10000,            // 1w tps / WindowSize
		MaxDelayCount:  1000,             // 1000 / WindowSize
		MaxFailedCount: 100,              // 100 / WindowSize
		MaxConnections: 10000,            // 10000 connections
		Retry:          30 * time.Second, // 30s
		MaxDelay:       1 * time.Second,  // 1s
	}
}

func (b *breaker) NewConnection(c Connection) error {
	defer c.Close()
	s := b.addConnection()
	switch s {
	case OPEN:
		return errors.New("Service Unavailable")
	case CLOSED:
		c.Serve()
		atomic.AddInt64(&b.c, -1)
	}
	return nil
}

func (b *breaker) NewRequest(req Request) error {
	s := b.currState()
	switch s {
	case OPEN:
		return errors.New("Service Unavailable")
	case CLOSED:
		p := b.currPosition()
		if p < 0 {
			return errors.New("Service Unavailable")
		}
		b.addEvent(p)
		t := time.Now()
		if err := req.Response(); err != nil {
			b.addFailedEvent(p)
		}
		if time.Now().Sub(t) > b.MaxDelay {
			b.addDealyEvent(p)
		}
	case HALFOPEN:
		t := time.Now()
		if err := req.Response(); err != nil || time.Now().Sub(t) > b.MaxDelay {
			break
		}
		b.s = CLOSED
		b.t = time.Now().Unix()
		for i, j := 0, b.WindowSize; i < j; i++ {
			b.ws[i].cnt = 0
			b.ws[i].fcnt = 0
			b.ws[i].dcnt = 0
		}
	}
	return nil
}

func (b *breaker) currState() int64 {
	b.Lock()
	defer b.Unlock()
	if b.s == OPEN {
		if t := time.Now(); t.Sub(b.opened) > b.Retry {
			b.opened = t
			return HALFOPEN
		}
	}
	return b.s
}

func (b *breaker) addConnection() int64 {
	b.Lock()
	defer b.Unlock()
	switch b.s {
	case OPEN:
		return OPEN
	case CLOSED:
		if b.c >= b.MaxConnections {
			b.s = OPEN
			b.opened = time.Now()
			return OPEN
		}
		b.c += 1
		return CLOSED
	}
	return CLOSED
}

func (b *breaker) currPosition() int {
	b.Lock()
	defer b.Unlock()
	t := int(time.Now().Unix() - b.t)
	switch {
	case t >= b.WindowSize:
		var cnt, fcnt, dcnt int64

		for i, j := 0, b.WindowSize; i < j; i++ {
			cnt += b.ws[i].cnt
			fcnt += b.ws[i].fcnt
			dcnt += b.ws[i].dcnt
		}
		if cnt >= b.MaxRequests || dcnt >= b.MaxDelayCount || fcnt >= b.MaxFailedCount {
			b.s = OPEN
			b.opened = time.Now()
			return -1
		}
		b.t = time.Now().Unix()
		for i, j := 0, b.WindowSize; i < j; i++ {
			b.ws[i].cnt = 0
			b.ws[i].fcnt = 0
			b.ws[i].dcnt = 0
		}
		return 0
	default:
		return t
	}
}

func (b *breaker) addEvent(pos int) {
	atomic.AddInt64(&b.ws[pos].cnt, 1)
}

func (b *breaker) addDealyEvent(pos int) {
	atomic.AddInt64(&b.ws[pos].fcnt, 1)
}

func (b *breaker) addFailedEvent(pos int) {
	atomic.AddInt64(&b.ws[pos].dcnt, 1)
}
