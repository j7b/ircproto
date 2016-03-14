package irc

import (
	"github.com/j7b/ircproto/irc/codes"
	"github.com/j7b/ircproto/irc/parser"
	"github.com/j7b/ircproto/events"
	"io"
	"log"
	"net/textproto"
)

var DEBUG = false

type Client struct {
	*textproto.Conn
	events  chan events.Event
	requeue chan events.Event
	buf     events.Event
	err     error
}

func (c *Client) Loop() (ok bool) {
	select {
	case c.buf, ok = <-c.requeue:
	case c.buf, ok = <-c.events:
	}
	return
}

func (c *Client) Event() events.Event {
	return c.buf
}

func (c *Client) Err() error {
	return c.err
}

func (c *Client) PrintfLine(f string, i ...interface{}) (err error) {
	if DEBUG {
		log.Printf("> "+f, i...)
	}
	return c.Conn.PrintfLine(f, i...)
}

func (c *Client) PrivMsg(target, msg string) (err error) {
	return c.PrintfLine("PRIVMSG %s :%s", target, msg)
}

func (c *Client) Notice(target, msg string) (err error) {
	return c.PrintfLine("NOTICE %s :%s", target, msg)
}

func (c *Client) Join(channel string, key ...string) (err error) {
	err = c.PrintfLine("JOIN %s", channel)
	if err == nil {
		for e := range c.events {
			switch t := e.(type) {
			default:
				c.requeue <- e
			case events.Join:
				return
			case events.Numeric:
				if codes.Code(t.Num()).OneOf(codes.ERR_NEEDMOREPARAMS,
					codes.ERR_BANNEDFROMCHAN,
					codes.ERR_INVITEONLYCHAN,
					codes.ERR_BADCHANNELKEY,
					codes.ERR_CHANNELISFULL,
					codes.ERR_BADCHANMASK,
					codes.ERR_NOSUCHCHANNEL,
					codes.ERR_TOOMANYCHANNELS,
					codes.ERR_TOOMANYTARGETS,
					codes.ERR_UNAVAILRESOURCE,
				) {
					err = Error{t}
					return
				}
			}
		}
	}
	return
}

func (c *Client) readloop() {
	if c.events != nil {
		panic("Event loop started out of sequence?")
	}
	c.events = make(chan events.Event, 128)
	c.requeue = make(chan events.Event, 128)
	go func() {
		for {
			s, err := c.ReadLine()
			if err != nil {
				if DEBUG {
					log.Println("<", err)
				}
				c.err = err
				close(c.events)
				return
			}
			if DEBUG {
				log.Println("<", s)
			}
			event := parser.Parse(s)
			if DEBUG {
				log.Printf("Parsed event %T", event)
			}
			switch t := event.(type) {
			case events.Ping:
				c.PrintfLine("PONG %s", t.Token())
			default:
				c.events <- event
			}
		}
	}()
}

type Error struct {
	events.Numeric
}

func (e Error) Error() string {
	return e.Payload()
}

func NewClient(con io.ReadWriteCloser, username string) (cl *Client, err error) {
	c := &Client{Conn: textproto.NewConn(con)}

	err = c.PrintfLine("NICK %s", username)
	if err != nil {
		return
	}
	err = c.PrintfLine("USER %s 0.0.0.0 0.0.0.0 %s", username, username)
	if err != nil {
		return
	}
	c.readloop()
	for e := range c.events {
		c.requeue <- e
		switch t := e.(type) {
		case events.Numeric:
			switch t.Num() {
			case 1:
				cl = c
				return
			case 465:
				err = Error{t}
				defer c.Close()
				return
			}
		}
	}
	err = c.Err()
	return
}
