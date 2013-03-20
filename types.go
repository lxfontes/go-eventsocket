package eventsocket

import (
	"bufio"
	"bytes"
	"net"
	"net/textproto"
	"sync"
	"time"
)

type ClientSettings struct {
	Address       string
	Reconnect     bool
	Timeout       time.Duration
	Password      string
	EventsChannel chan Event
}

type Client struct {
	Settings   *ClientSettings
	Reconnects int
	Connected  bool
	Connection net.Conn
	rw         *bufio.ReadWriter

	//channels for internal communication
	executeChan chan Event
	apiChan     chan Event
	lock        sync.Mutex
}

type Command struct {
	Lock bool
	Uuid string
	App  string
	Args string
}

type EventType int

const (
	EventError EventType = iota
	EventState
	EventAuth
	EventReply
	EventApi
	EventDisconnect
	EventGeneric
)

type Event interface {
	Type() EventType
	Success() bool
	Headers() *textproto.MIMEHeader
	Body() []byte
	String() string
}

type eventReply struct {
	realType EventType
	headers  textproto.MIMEHeader
	body     []byte
	success  bool
}

type connectionState struct {
	connected bool
}

func (c connectionState) String() string {
	if c.connected {
		return "connected"
	}
	return "not connected"
}

func (c connectionState) Success() bool {
	return c.connected
}

func (c connectionState) Type() EventType {
	return EventState
}

func (c connectionState) Headers() *textproto.MIMEHeader {
	return nil
}

func (c connectionState) Body() []byte {
	return nil
}

func (e *eventReply) Success() bool {
	return e.success
}

func (e *eventReply) Type() EventType {
	return e.realType
}

func (e *eventReply) Headers() *textproto.MIMEHeader {
	return &e.headers
}

func (e *eventReply) Body() []byte {
	return e.body
}

func (e *eventReply) String() string {
	switch e.realType {
	case EventError:
		return "error"
	}
	return e.headers.Get("Content-Type")

}

func (cmd *Command) GetExecute() []byte {
	bbuf := bytes.NewBufferString("sendmsg")
	if len(cmd.Uuid) > 0 {
		bbuf.WriteString(" ")
		bbuf.WriteString(cmd.Uuid)
	}
	bbuf.Write(singleLine)
	bbuf.WriteString("call-command: execute")
	bbuf.Write(singleLine)
	bbuf.WriteString("execute-app-name: ")
	bbuf.WriteString(cmd.App)
	bbuf.Write(singleLine)
	if len(cmd.Args) > 0 {
		bbuf.WriteString("execute-app-arg: ")
		bbuf.WriteString(cmd.Args)
		bbuf.Write(singleLine)
	}
	if cmd.Lock {
		bbuf.WriteString("event-lock: true")
		bbuf.Write(singleLine)
	}
	bbuf.Write(doubleLine)
	return bbuf.Bytes()
}

func (cmd *Command) GetSend() []byte {
	bbuf := bytes.NewBufferString(cmd.App)
	bbuf.WriteString(" ")
	bbuf.WriteString(cmd.Args)
	bbuf.Write(singleLine)
	bbuf.Write(doubleLine)
	return bbuf.Bytes()
}
