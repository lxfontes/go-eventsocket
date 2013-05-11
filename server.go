package eventsocket

import (
	"bufio"
	"bytes"
	"net"
)

func CreateServer(settings *ServerSettings) (*Server, error) {
	var err error
	retServer := new(Server)
	retServer.Settings = settings
	retServer.EventsChannel = make(chan *Event)
	//bind and go listen for clients
	retServer.Listener, err = net.Listen("tcp", retServer.Settings.Address)
	if err != nil {
		return nil, err
	}
	go retServer.loop()
	return retServer, nil
}

func (scon *ServerConnection) Close() {
	scon.eslCon.Close()
}

func (scon *ServerConnection) loop() {
	//authenticate socket
	cBuf := bytes.NewBufferString("connect")
	cBuf.Write(doubleLine)
	scon.rw.Write(cBuf.Bytes())
	err := scon.rw.Flush()

	if err != nil {
		scon.eslCon.Close()
		return
	}

	channelData := readMessage(scon.rw)

	if channelData.Type != EventReply {
		scon.eslCon.Close()
		return
	}

	scon.Connection.ChannelData = channelData.Headers
	scon.Server.EventsChannel <- &Event{Success: true, Type: EventState, Connection: &scon.Connection}
	scon.Connected = true

	for scon.Connected {
		message := readMessage(scon.rw)

		switch message.Type {
		case EventError:
			//disconnect
			scon.eslCon.Close()
			scon.Connected = false
			scon.Server.EventsChannel <- &Event{Success: false, Type: EventState, Connection: &scon.Connection}
		case EventDisconnect:
			//disconnect
			scon.Connected = false
			scon.Server.EventsChannel <- &Event{Success: false, Type: EventState, Connection: &scon.Connection}
		case EventReply, EventApi:
			message.Connection = &scon.Connection
			scon.apiChan <- message
		case EventGeneric:
			message.Connection = &scon.Connection
			scon.Server.EventsChannel <- message
		}

	}
}

func (server *Server) loop() {
	for {
		conn, err := server.Listener.Accept()
		if err != nil {
			//what to do here?
			return
		}
		scon := new(ServerConnection)
		scon.apiChan = make(chan *Event)
		scon.Server = server
		scon.eslCon = conn
		scon.rw = bufio.NewReadWriter(
			bufio.NewReaderSize(scon.eslCon, BufferSize),
			bufio.NewWriter(scon.eslCon))
		go scon.loop()
	}
}
