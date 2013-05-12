package eventsocket

import (
	"bufio"
	"bytes"
	"fmt"
	_ "fmt"
	"net"
	"sync"
	"time"
)

type ServerListener interface {
	OnNewConnection(con *Connection)
}

type ConnectionListener interface {
	OnConnect(con *Connection)
	OnEvent(con *Connection, evt *Event)
	OnDisconnect(con *Connection, msg *Event)
	OnClose(con *Connection)
}

type ServerSettings struct {
	Address  string
	Listener ServerListener
}

type Connection struct {
	//server specific
	Server *Server

	//client specific
	Reconnects  int
	ChannelData ESLkv
	Connected   bool

	//generic
	Listener ConnectionListener
	eslCon   net.Conn
	rw       *bufio.ReadWriter
	lock     sync.Mutex
	apiChan  chan *Event
}

type Server struct {
	Settings    ServerSettings
	NetListener net.Listener
	EvListener  ServerListener
}

/* Connection */

func (con *Connection) Send(cmd string, args ...string) (*Event, error) {
	bbuf := bytes.NewBufferString(cmd)
	for _, item := range args {
		bbuf.WriteString(" ")
		bbuf.WriteString(item)
	}
	bbuf.Write(doubleLine)

	//all commands block until FS replies
	con.lock.Lock()
	defer con.lock.Unlock()

	sendBytes(con.rw, bbuf.Bytes())
	evt := <-con.apiChan

	return evt, nil
}

func (con *Connection) Event(cmd string, headers map[string]string, body []byte) (*Event, error) {
	bbuf := bytes.NewBufferString("sendevent ")
	bbuf.WriteString(cmd)
	bbuf.Write(singleLine)

	for k, v := range headers {
		bbuf.WriteString(k)
		bbuf.WriteString(": ")
		bbuf.WriteString(v)
		bbuf.Write(singleLine)
	}

	ll := len(body)
	lf := fmt.Sprintf("content-length: %d%s", ll, doubleLine)
	bbuf.WriteString(lf)

	bbuf.Write(body)

	//all commands block until FS replies
	con.lock.Lock()
	defer con.lock.Unlock()

	sendBytes(con.rw, bbuf.Bytes())
	evt := <-con.apiChan

	return evt, nil
}

func (con *Connection) Execute(cmd *Command) (*Event, error) {
	con.lock.Lock()
	defer con.lock.Unlock()

	sendBytes(con.rw, cmd.GetExecute())
	evt := <-con.apiChan

	return evt, nil
}
func (con *Connection) tryConnect(client *Client) bool {
	var err error
	con.eslCon, err = net.DialTimeout("tcp", client.Settings.Address, client.Settings.Timeout)

	if err != nil {
		return false
	}

	con.rw = bufio.NewReadWriter(
		bufio.NewReaderSize(con.eslCon, BufferSize),
		bufio.NewWriter(con.eslCon))

	auth := readMessage(con.rw)
	if auth.Type != EventAuth {
		con.eslCon.Close()
		return false
	}

	pwBuf := bytes.NewBufferString("auth ")
	pwBuf.WriteString(client.Settings.Password)
	pwBuf.Write(doubleLine)
	con.eslCon.Write(pwBuf.Bytes())
	err = con.rw.Flush()

	if err != nil {
		con.eslCon.Close()
		return false
	}

	authResp := readMessage(con.rw)

	if authResp.Type != EventReply || !authResp.Success {
		con.eslCon.Close()
		return false
	}

	con.Connected = true
	con.Reconnects += 1
	return true
}

func (con *Connection) Close() {
	con.Connected = false
	con.Listener.OnClose(con)
	con.eslCon.Close()
}

func (con *Connection) setupConnect() bool {
	//authenticate socket
	cBuf := bytes.NewBufferString("connect")
	cBuf.Write(doubleLine)
	con.rw.Write(cBuf.Bytes())
	err := con.rw.Flush()

	if err != nil {
		con.eslCon.Close()
		return false
	}

	channelData := readMessage(con.rw)

	if channelData.Type != EventReply {
		con.eslCon.Close()
		return false
	}

	con.ChannelData = channelData.Headers
	con.Connected = true
	return true
}

func (con *Connection) Loop() {
	//spin this on a separate goroutine so we can start handling api events
	//people will likely subscribe to events here
	go con.Listener.OnConnect(con)
	for con.Connected {
		message := readMessage(con.rw)

		switch message.Type {
		case EventError:
			//disconnect
			con.Close()
		case EventDisconnect:
			//disconnect
			con.Listener.OnDisconnect(con, message)
		case EventReply, EventApi:
			con.apiChan <- message
		case EventGeneric:
			con.Listener.OnEvent(con, message)
		}

	}
}

/* Server */

func CreateServer(settings ServerSettings) (*Server, error) {
	var err error
	retServer := new(Server)
	retServer.Settings = settings
	retServer.EvListener = settings.Listener
	//bind and go listen for clients
	retServer.NetListener, err = net.Listen("tcp", retServer.Settings.Address)
	if err != nil {
		return nil, err
	}
	return retServer, nil
}
func (server *Server) Loop() {
	for {
		conn, err := server.NetListener.Accept()
		if err != nil {
			//what to do here?
			return
		}
		scon := new(Connection)
		scon.apiChan = make(chan *Event)
		scon.Server = server
		scon.eslCon = conn
		scon.rw = bufio.NewReadWriter(
			bufio.NewReaderSize(scon.eslCon, BufferSize),
			bufio.NewWriter(scon.eslCon))
		if scon.setupConnect() {
			server.EvListener.OnNewConnection(scon)
		}
	}
}

/* CLIENT */

type ClientSettings struct {
	Address  string
	Timeout  time.Duration
	Password string
	Listener ConnectionListener
}
type Client struct {
	Settings   ClientSettings
	Connected  bool
	Connection Connection
}

func CreateClient(settings ClientSettings) (*Client, error) {
	retClient := new(Client)
	retClient.Settings = settings
	retClient.Connection.Listener = settings.Listener
	retClient.Connection.apiChan = make(chan *Event)
	return retClient, nil
}

func (client *Client) Loop() {
	for {
		if !client.Connected {
			client.Connected = client.Connection.tryConnect(client)
			if !client.Connected {
				time.Sleep(client.Settings.Timeout)
			}
			continue
		}
		client.Connection.Loop()
		client.Connected = false
	}
}
