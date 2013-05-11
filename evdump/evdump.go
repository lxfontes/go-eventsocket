package main

import (
	"flag"
	"fmt"
	esl "github.com/lxfontes/go-eventsocket"
	"time"
)

var address *string = flag.String("address", "localhost:8021", "Freeswitch Event Socket Address")
var password *string = flag.String("password", "cluecon", "Event Socket Password")

type Dumper struct{}

func (n *Dumper) OnConnect(con *esl.Connection) {
	fmt.Println("Connected")
	evt, err := con.Send("event", "plain", "ALL")
	if err != nil {
		fmt.Println("ahn?")
		return
	}
	fmt.Println(evt.Success)
}

func (n *Dumper) OnDisconnect(con *esl.Connection) {
	fmt.Println("disconnect")
}

func (n *Dumper) OnEvent(con *esl.Connection, evt *esl.Event) {
	fmt.Println(string(evt.Body))
	fmt.Println(evt.EventBody.Get("Event-Date-Local"))
}

func (n *Dumper) OnClose(con *esl.Connection) {
	fmt.Println("close")
}

func main() {

	flag.Parse()

	fmt.Println("Connecting to", *address, "using password", *password)
	client, err := esl.CreateClient(esl.ClientSettings{
		Address:  *address,
		Password: *password,
		Timeout:  10 * time.Second,
		Listener: new(Dumper),
	})

	if err != nil {
		fmt.Println("Something went wrong ", err)
	}

	client.Loop()
}
