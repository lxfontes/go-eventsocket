package eventsocket

import (
	"fmt"
	"testing"
	"time"
)

type Nooper struct{}

func (n *Nooper) OnConnect(con *Connection) {
	fmt.Println("Connected")
	evt, err := con.Send("event", "plain", "HEARTBEAT")
	if err != nil {
		fmt.Println("ahn?")
		return
	}
	fmt.Println(evt.Success)
}

func (n *Nooper) OnDisconnect(con *Connection) {
	fmt.Println("disconnect")
}

func (n *Nooper) OnEvent(con *Connection, evt *Event) {
	fmt.Println("event", evt.EventBody.Get("Event-Name"))
	con.Send("quit")
}

func (n *Nooper) OnClose(con *Connection) {
	fmt.Println("close")
}

func Test_Client(t *testing.T) {
	t.Log("Test Basic Client")

	client, err := CreateClient(ClientSettings{
		Address:  "localhost:8021",
		Password: "fongopass",
		Timeout:  1 * time.Second,
		Listener: new(Nooper),
	})

	if err != nil {
		fmt.Println("Something went wrong ", err)
	}
	if client != nil {
	}

	//	client.Loop()

}
