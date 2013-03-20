package eventsocket

import (
	"fmt"
	"testing"
	"time"
)

func Test_BasicConn(t *testing.T) {
	t.Log("Test Basic Client")

	ch := make(chan Event)
	client, err := CreateClient(&ClientSettings{
		Address:       "localhost:8021",
		Password:      "fongopass",
		Timeout:       10 * time.Second,
		Reconnect:     true,
		EventsChannel: ch,
	})

	if err != nil {
		fmt.Println("Something went wrong ", err)
	}

	for {
		msg := <-ch

		switch {
		case msg.Type() == EventState && msg.Success():
			fmt.Println("Connected")
			resp, err := client.Subscribe("ALL")
			fmt.Println(resp, err)
			resp, err = client.Api(&Command{
				App:  "api",
				Args: "reloadxml",
			})
			fmt.Println(resp, err)
		case msg.Type() == EventGeneric:
			fmt.Println(msg)
			fmt.Println(string(msg.Body()))
		}
	}
	client.Close()
}
