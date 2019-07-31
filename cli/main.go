package main

import (
	"bufio"
	"fmt"
	"log"
	"net/textproto"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/nnsgmsone/protocol"
	"github.com/nnsgmsone/units/unit"
)

type dealFunc (func(*textproto.Writer, []string))

var u unit.Unit

var dealRegister map[string]dealFunc

func init() {
	dealRegister = make(map[string]dealFunc)

	dealRegister["help"] = deal0

	dealRegister["connect"] = deal1
	dealRegister["disconnect"] = deal2

	dealRegister["delRoom"] = deal3
	dealRegister["addRoom"] = deal4
	dealRegister["chgRoom"] = deal5

	dealRegister["exit"] = deal99
}

func main() {
	go func() {
		for {
			ch := make(chan os.Signal)
			signal.Notify(ch)
			sig := <-ch
			if sig.String() == "quit" || sig.String() == "killed" || sig.String() == "interrupt" {
				if u != nil {
					u.Close()
				}
				os.Exit(0)
			}
		}
	}()
	run(textproto.NewReader(bufio.NewReader(os.Stdin)), bufio.NewWriter(os.Stdout))
}

func run(r *textproto.Reader, w *bufio.Writer) {
	for {
		w.WriteString("# ")
		w.Flush()
		s, err := r.ReadLine()
		if err != nil {
			if u != nil {
				u.Close()
			}
			log.Fatal(err)
		}
		if len(s) == 0 {
			continue
		}
		xs := strings.Fields(s)
		f, ok := dealRegister[string(xs[0])]
		if !ok {
			fmt.Printf("Illegal Command: '%v'\n", xs)
			continue
		}
		f(textproto.NewWriter(w), xs[1:])
	}
}

func deal0(w *textproto.Writer, args []string) {
	w.PrintfLine("\thelp")
	w.PrintfLine("\texit")
	w.PrintfLine("\tconnect [manager' address]")
	w.PrintfLine("\tdisconnect")
	w.PrintfLine("\tdelRoom [room's name]")
	w.PrintfLine("\taddRoom [room's name] [room's number]")
	w.PrintfLine("\tchgRoom [room's name] [room's number]")
}

// connect [manager' address]
func deal1(w *textproto.Writer, args []string) {
	if len(args) < 1 {
		w.PrintfLine("Usage: connect [manager's address]")
		return
	}
	if u != nil {
		u.Close()
	}
	u = unit.New(args[0], 2*time.Second)
	w.PrintfLine("connect ok")
}

// disconnect
func deal2(w *textproto.Writer, args []string) {
	if u != nil {
		u.Close()
		u = nil
	}
	w.PrintfLine("disconnect ok")
}

// delRoom name
func deal3(w *textproto.Writer, args []string) {
	if len(args) < 1 {
		w.PrintfLine("Usage: delRoom [room's name]")
		return
	}
	if u == nil {
		w.PrintfLine("Please Connect First")
		return
	}
	msg, err := u.SendAndRecv("", protocol.DelRoom(args[0]))
	if err != nil {
		w.PrintfLine(err.Error())
		return
	}
	switch m := msg.Msg.(type) {
	case protocol.MessageError:
		w.PrintfLine(m.M)
	case protocol.MessageSlice:
		w.PrintfLine(m.M)
	case protocol.MessageInteger:
		w.PrintfLine(fmt.Sprintf("%d", m.M))
	default:
		w.PrintfLine("server error")
	}
}

// addRoom name number
func deal4(w *textproto.Writer, args []string) {
	if len(args) < 2 {
		w.PrintfLine("Usage: addRoom [room's name] [room's number]")
		return
	}
	if u == nil {
		w.PrintfLine("Please Connect First")
		return
	}
	msg, err := u.SendAndRecv("", protocol.AddRoom(args[0], args[1]))
	if err != nil {
		w.PrintfLine(err.Error())
		return
	}
	switch m := msg.Msg.(type) {
	case protocol.MessageError:
		w.PrintfLine(m.M)
	case protocol.MessageSlice:
		w.PrintfLine(m.M)
	case protocol.MessageInteger:
		w.PrintfLine(fmt.Sprintf("%d", m.M))
	default:
		w.PrintfLine("server error")
	}
}

// chgRoom name number
func deal5(w *textproto.Writer, args []string) {
	if len(args) < 2 {
		w.PrintfLine("Usage: chgRoom [room's name] [room's number]")
		return
	}
	if u == nil {
		w.PrintfLine("Please Connect First")
		return
	}
	msg, err := u.SendAndRecv("", protocol.ChgRoom(args[0], args[1]))
	if err != nil {
		w.PrintfLine(err.Error())
		return
	}
	switch m := msg.Msg.(type) {
	case protocol.MessageError:
		w.PrintfLine(m.M)
	case protocol.MessageSlice:
		w.PrintfLine(m.M)
	case protocol.MessageInteger:
		w.PrintfLine(fmt.Sprintf("%d", m.M))
	default:
		w.PrintfLine("server error")
	}
}

// exit
func deal99(w *textproto.Writer, args []string) {
	if u != nil {
		u.Close()
	}
	os.Exit(0)
}
