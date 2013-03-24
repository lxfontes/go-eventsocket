package eventsocket

import (
	"bufio"
	"bytes"
	"net"
	"net/textproto"
	"sync"
	"time"
)

const BufferSize = 16 * 1024

type ClientSettings struct {
	Address  string
	Timeout  time.Duration
	Password string
}

type Client struct {
	Settings      *ClientSettings
	EventsChannel chan *Event
	Reconnects    int
	connection
}

//Non exported struct, gets exposed via composition
type connection struct {
	//available while in Server Mode
	ChannelData textproto.MIMEHeader
	//true if freeswitch is still able to receive commands (ex: prior disconnect notice)
	Connected bool
	eslCon    net.Conn
	rw        *bufio.ReadWriter
	lock      sync.Mutex
	//channels for internal communication
	apiChan chan *Event
}

type ServerSettings struct {
	Address string
}

type ServerConnection struct {
	connection
	Server *Server
}

type Server struct {
	Settings      *ServerSettings
	EventsChannel chan *Event
	Listener      net.Listener
}

type Command struct {
	Lock bool
	Uuid string
	App  string
	Args string
}

type JsonBody map[string]string

// EventType indicates how/why this event is being raised
type EventType int

const (
	//Parser and Socket errors
	EventError EventType = iota
	//Client Mode: Connection established / terminated
	//Server Mode: New connection / Lost connection
	EventState
	//Library internal, used during auth phase
	EventAuth
	//Command Reply
	EventReply
	//API (reloadxml,reloadacl) commands
	EventApi
	//Received a disconnect notice (linger might still be on)
	EventDisconnect
	//Subscribed events
	EventGeneric
)

type Event struct {
	Type       EventType
	Headers    textproto.MIMEHeader
	Body       []byte
	Connection *connection
	Success    bool
}

// GetExecute formats a dialplan command to be sent over TCP connection
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

func (evt *Event) String() string {
	evtype := evt.Headers.Get("Content-Type")
	bbuf := bytes.NewBufferString(evtype)
	bbuf.WriteString(" ")
	return bbuf.String()
}
