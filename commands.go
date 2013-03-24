package eventsocket

import (
	"fmt"
)

// TODO: A few helpers to well....'help' sending dialplan commands
// They block in some situations (ex: ringready), but not others (ex: bridge)

func (con *connection) Answer() (*Event, error) {
	cmd := Command{
		App:  "answer",
		Lock: true,
	}
	return con.Execute(&cmd)
}

func (con *connection) VerboseEvents() (*Event, error) {
	cmd := Command{
		App:  "verbose_events",
		Lock: true,
	}
	return con.Execute(&cmd)
}

func (con *connection) Bridge(endpoint string, lock bool) (*Event, error) {
	cmd := Command{
		App:  "bridge",
		Args: endpoint,
		Lock: lock,
	}
	return con.Execute(&cmd)
}

func (con *connection) Hangup(reason string) (*Event, error) {
	cmd := Command{
		App:  "hangup",
		Args: reason,
		Lock: true,
	}
	return con.Execute(&cmd)
}

func (con *connection) RingReady() (*Event, error) {
	cmd := Command{
		App:  "ring_ready",
		Lock: false,
	}
	return con.Execute(&cmd)
}

func (con *connection) RecordSession(filename string) (*Event, error) {
	cmd := Command{
		App:  "record_session",
		Args: filename,
		Lock: false,
	}
	return con.Execute(&cmd)
}

func (con *connection) Read(min int, max int, audio string, variable string, timeout int, terminators string) (*Event, error) {
	args := fmt.Sprint("%d %d %s %s %d %s", min, max, audio, variable, timeout, terminators)
	cmd := Command{
		App:  "read",
		Args: args,
		Lock: false,
	}
	return con.Execute(&cmd)
}

func (con *connection) Set(key string, value string) (*Event, error) {
	args := fmt.Sprint("%s=%s", key, value)
	cmd := Command{
		App:  "read",
		Args: args,
		Lock: false,
	}
	return con.Execute(&cmd)
}

func (con *connection) UnSet(key string) (*Event, error) {

	cmd := Command{
		App:  "read",
		Args: key,
		Lock: false,
	}
	return con.Execute(&cmd)
}

func (con *connection) Exit() (*Event, error) {
	return con.Send("exit")
}

func (con *connection) EventPlain(event string) (*Event, error) {
	return con.Send("event", "plain", event)
}

func (con *connection) EventJson(event string) (*Event, error) {
	return con.Send("event", "json", event)
}

func (con *connection) EventXml(event string) (*Event, error) {
	return con.Send("event", "xml", event)
}

func (con *connection) Linger() (*Event, error) {
	return con.Send("linger")
}

func (con *connection) Filter(key string, value string) (*Event, error) {
	return con.Send("filter", key, value)
}

func (con *connection) FilterDelete(key string, value string) (*Event, error) {
	return con.Send("filter", "delete", key, value)
}
