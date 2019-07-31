package breaker

import (
	"sync"
	"time"
)

const (
	OPEN = iota
	CLOSED
	HALFOPEN
)

type Request interface {
	Response() error
}

type Connection interface {
	Serve() error
	Close() error
}

type Breaker interface {
	NewRequest(Request) error
	NewConnection(Connection) error
}

type Config struct {
	WindowSize     int           // window大小，单位为s
	MaxConnections int64         // 最大链接数目
	MaxRequests    int64         // window时间内的最大请求数
	MaxDelayCount  int64         // window时间内的超过延迟的最大请求数
	MaxFailedCount int64         // window时间内的连续失败次数
	Retry          time.Duration // 开启后重试时间
	MaxDelay       time.Duration // 单个请求最大延迟
}

type window struct {
	cnt  int64 // 请求数目
	fcnt int64 // 失败的请求数目
	dcnt int64 // 延迟过大的请求数目
}

type breakerConnection struct {
	sync.Mutex
	cnt int64 // 链接数目
}

type breaker struct {
	*Config
	sync.Mutex
	c      int64 // 链接数目
	t      int64 // 上一次命中的窗口时间
	s      int64 // 状态
	ws     []window
	opened time.Time
}
