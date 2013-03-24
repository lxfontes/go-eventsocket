package eventsocket

import (
	"bufio"
	"bytes"
	"fmt"
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
		fmt.Println("Closing due to", channelData.Type)
		scon.eslCon.Close()
		return
	}

	scon.Server.EventsChannel <- &Event{Success: true, Type: EventState, Connection: &scon.connection}

	for {
		message := readMessage(scon.rw)

		switch message.Type {
		case EventError:
			//disconnect
			scon.eslCon.Close()
			scon.Connected = false
			scon.Server.EventsChannel <- &Event{Success: false, Type: EventState, Connection: &scon.connection}
		case EventDisconnect:
			//disconnect
			scon.eslCon.Close()
			scon.Connected = false
			scon.Server.EventsChannel <- &Event{Success: false, Type: EventState, Connection: &scon.connection}
		case EventReply, EventApi:
			scon.apiChan <- message
		case EventGeneric:
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
