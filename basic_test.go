package eventsocket

import (
	"fmt"
	"testing"
	"time"
)

type HandlerTest struct{}

func (h *HandlerTest) CreateConnection(conn *FSConnection) bool {
	fmt.Println("Will connect")
	return true
}

func (h *HandlerTest) ConnectionAccepted(conn *FSConnection) {
	fmt.Println("Connected")
	conn.EventPlain("ALL")
}

func (h *HandlerTest) CloseConnection(conn *FSConnection) {
	fmt.Println("Lost connection")
}

func (h *HandlerTest) HandleEvent(conn *FSConnection, evt *FSMessage) {
	fmt.Println("Received event")
}

func Test_BasicConn(t *testing.T) {
	t.Log("Creating new connection")
	var h = new(HandlerTest)
	client, err := CreateClient("localhost:8021", "fongopass", h, 10*time.Second)
	if err != nil {
		fmt.Println("Something went wrong ", err)
	}

	go client.Loop()

	for {
		time.Sleep(1 * time.Second)
	}

}
