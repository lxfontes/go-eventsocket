package eventsocket

import (
	"bufio"
	"bytes"
	_ "fmt"
	"net"
	"time"
)

func CreateClient(settings *ClientSettings) (*Client, error) {
	retClient := new(Client)
	retClient.Settings = settings
	retClient.EventsChannel = make(chan *Event)
	retClient.apiChan = make(chan *Event)
	go retClient.loop()
	return retClient, nil
}

func (client *Client) tryConnect() {
	var err error
	client.eslCon, err = net.DialTimeout("tcp", client.Settings.Address, client.Settings.Timeout)

	if err != nil {
		time.Sleep(client.Settings.Timeout)
		return
	}

	//Got Connection, initiate authentication
	client.rw = bufio.NewReadWriter(
		bufio.NewReaderSize(client.eslCon, BufferSize),
		bufio.NewWriter(client.eslCon))

	//wait for authentication request
	auth := readMessage(client.rw)
	if auth.Type != EventAuth {
		//invalid handshake
		return
	}

	//send password
	pwBuf := bytes.NewBufferString("auth ")
	pwBuf.WriteString(client.Settings.Password)
	pwBuf.Write(doubleLine)
	client.rw.Write(pwBuf.Bytes())
	err = client.rw.Flush()

	if err != nil {
		client.eslCon.Close()
		return
	}

	authResp := readMessage(client.rw)

	if authResp.Type != EventReply {
		client.eslCon.Close()
		return
	}

	if !authResp.Success {
		client.eslCon.Close()
		return
	}

	//connected and authed
	client.Connected = true
	client.Reconnects += 1
	client.EventsChannel <- &Event{Success: true, Type: EventState, Connection: &client.Connection}
}

// Loop communicates with main process via Handler's message channel
// To be called after authentication or new Connection
func (client *Client) loop() {
	for {
		if !client.Connected {
			client.tryConnect()
			continue
		}

		message := readMessage(client.rw)
		message.Connection = &client.Connection

		switch message.Type {
		case EventError:
			//disconnect
			client.eslCon.Close()
			client.Connected = false
			client.EventsChannel <- &Event{Success: false, Type: EventState, Connection: &client.Connection}
		case EventDisconnect:
			//disconnect
			client.eslCon.Close()
			client.Connected = false
			client.EventsChannel <- &Event{Success: false, Type: EventState, Connection: &client.Connection}
		case EventReply, EventApi:
			client.apiChan <- message
		case EventGeneric:
			client.EventsChannel <- message
		}

	}
}
