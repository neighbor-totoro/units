package kventry

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/nnsgmsone/protocol"
	"github.com/nnsgmsone/units/relay"
	"github.com/valyala/fasthttp"
)

func New(ry relay.Relay, port int, hu, name string) *kventry {
	ch := make(chan struct{})
	if mch, err := ry.Join(name, 1024, 4*time.Second); err != nil {
		return nil
	} else {
		return &kventry{port, hu, ry, ch, mch}
	}
}

func (kv *kventry) Run() {
	h := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/set":
			dealRequest(S0, kv, ctx)
		case "/del":
			dealRequest(S1, kv, ctx)
		case "/get":
			dealRequest(S2, kv, ctx)
		default:
			ctx.Error("Unsupport Path", fasthttp.StatusNotFound)
		}
	}
	go kv.ry.Run()
	go kv.dealMessage()
	fasthttp.ListenAndServe(fmt.Sprintf(":%v", kv.port), h)
}

func (kv *kventry) Stop() {
	kv.ry.Stop()
	kv.ch <- struct{}{}
	<-kv.ch
	close(kv.ch)
}

func (kv *kventry) dealMessage() {
	for {
		select {
		case <-kv.ch:
			kv.ch <- struct{}{}
			return
		case msg := <-kv.mch:
			log.Printf("Recv: %v\n", msg.M)
		}
	}
}

func dealRequest(typ int, kv *kventry, ctx *fasthttp.RequestCtx) {
	var err error
	var data []byte
	var k, v string
	var cmd []string
	var hr HttpResult
	var msg *relay.Message
	var mp map[string]interface{}

	ctx.Response.SetStatusCode(200)
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Content-Type", "application/json")
	if err = json.Unmarshal(ctx.PostBody(), &mp); err != nil {
		ctx.Response.SetStatusCode(400)
		goto ERR
	}
	switch typ {
	case S0: // set k v
		if k, err = getString(mp, "key"); err != nil {
			ctx.Response.SetStatusCode(400)
			goto ERR
		}
		if v, err = getString(mp, "value"); err != nil {
			ctx.Response.SetStatusCode(400)
			goto ERR
		}
		cmd = set(k, v)
	case S1: // del k
		if k, err = getString(mp, "key"); err != nil {
			ctx.Response.SetStatusCode(400)
			goto ERR
		}
		cmd = del(k)
	case S2: // get k
		if k, err = getString(mp, "key"); err != nil {
			ctx.Response.SetStatusCode(400)
			goto ERR
		}
		cmd = get(k)
	}
	msg = &relay.Message{protocol.NewMessage(kv.hu, cmd), make(chan *protocol.Message)}
	kv.ry.Channel() <- msg
	msg.M = <-msg.Ch // reuse
	switch m := msg.M.Msg.(type) {
	case protocol.MessageError:
		hr.Err = m.M
	case protocol.MessageArray:
		hr.Msg = m.M[0]
	default:
		hr.Err = fmt.Sprintf("Unpredictable Error")
	}
	data, _ = json.Marshal(hr)
	ctx.Write(data)
	return
ERR:
	hr.Err = err.Error()
	data, _ = json.Marshal(hr)
	ctx.Write(data)
}

// get key
func get(k string) []string {
	return []string{"get", k}
}

// del key
func del(k string) []string {
	return []string{"del", k}
}

// set key value
func set(k, v string) []string {
	return []string{"set", k, v}
}

func getString(mp map[string]interface{}, k string) (string, error) {
	{
		fmt.Printf("k: %v; mp: %v\n", k, mp)
	}
	v, ok := mp[k]
	if !ok {
		return "", errors.New("Not Exist")
	}
	if _, ok := v.(string); !ok {
		return "", errors.New("Not String")
	}
	return v.(string), nil
}
