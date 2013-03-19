package eventsocket

import (
	"bufio"
	"net"
	"net/textproto"
	"sync"
	"time"
)

type FSHandler interface {
	// When in server mode, dictates if a connection should be accepted or not
	// When in client mode, dictates if a connection should be opened or not
	CreateConnection(*FSConnection) bool

	// When connection is established
	ConnectionAccepted(*FSConnection)

	// Called once Freeswitch disconnects (gracefully or not)
	CloseConnection(*FSConnection)
	// Unbound events
	HandleEvent(*FSConnection, *FSMessage)
}

type FSListener struct {
	Handler    FSHandler
	Timeout    time.Duration
	Connection net.Conn
	Address    string
	ClientMode bool
	// Available on outbound sockets (Go->FS)
	Password string
	QuitChan chan int
}

type FSConnection struct {
	Listener   *FSListener
	Connection net.Conn
	Connected  bool
	// Available for inbound sockets (FS->Go)
	ChannelData textproto.MIMEHeader
	rw          *bufio.ReadWriter
	lock        sync.Mutex
}

type FSMessageType int

const (
	ParserError FSMessageType = iota
	RequestAuthentication
	CommandReply
	EventPlain
	DisconnectNotice
)

type FSMessage struct {
	Type    FSMessageType
	Success bool
	Headers textproto.MIMEHeader
	Body    textproto.MIMEHeader
}
