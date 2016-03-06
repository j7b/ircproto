package ircproto

import (
	"fmt"
	"log"
	"net/textproto"
	"strings"
)

var DEBUG = false

type msg struct {
	text string
}

type msgping struct {
	*msg
}

func (m *msg) Text() string {
	return m.text
}

func (m *msgping) Ping() (tok string) {
	parts := strings.Split(m.text, " ")
	tok = parts[1]
	return
}

type Message interface {
	Text() string
}

type Ping interface {
	Message
	Ping() string
}

type ircchannel struct {
	name   string
	topic  string
	lusers []string
}

type textchan chan string

func (t textchan) close() {
	defer func() {
		recover()
	}()
	close(t)
}

func (t textchan) Printf(format string, i ...interface{}) (err error) {
	defer func() {
		if recover() != nil {
			err = fmt.Errorf("Channel closed")
		}
	}()
	t <- fmt.Sprintf(format, i...)
	return
}

func (t textchan) pong(s string) error {
	return t.Printf("PONG %s", s)
}

type msgchan chan Message

func (m msgchan) close() {
	defer func() {
		recover()
	}()
	close(m)
}
func (m msgchan) requeue(mess Message) {
	go func() {
		m <- mess
	}()
}
func parse(line string) (m Message) {
	if DEBUG {
		log.Println("parse", line)
	}
	msg := &msg{line}
	switch {
	case strings.Index(line, "PING ") == 0:
		if DEBUG {
			log.Println("PING", line)
		}
		return &msgping{msg}
	default:
	}
	return msg
}

type client struct {
	con     *textproto.Conn
	inbound msgchan
	textchan
}

func (c *client) Quit(msg string) (err error) {
	err = c.Printf("QUIT %s", msg)
	if err == nil {
		defer c.textchan.close()
		defer c.inbound.close()
	}
	return
}

func (c *client) inbounds() {
	for m := range c.inbound {
		if DEBUG {
			log.Println("Inbound msg", m)
		}
		switch t := m.(type) {
		case Ping:
			if DEBUG {
				log.Println("Is ping")
			}
			c.pong(t.Ping())
		default:
			if DEBUG {
				log.Println("Is default")
			}
		}
	}
}

func (c *client) outbounds() {
	defer close(c.textchan)
	for l := range c.textchan {
		if DEBUG {
			log.Println(">>", l)
		}
		if err := c.con.PrintfLine(l); err != nil {
			return
		}
	}
}

func New(addr, nick, realname string) (c *client, err error) {
	con, err := textproto.Dial("tcp", addr)
	if err != nil {
		return
	}
	err = con.PrintfLine("NICK %s", nick)
	if err != nil {
		return
	}
	err = con.PrintfLine("USER %s 0.0.0.0 0.0.0.0 %s", nick, realname)
	inbound := make(msgchan, 256)
	outbound := make(textchan, 256)
	go func() {
		defer close(inbound)
		var line string
		var err error
		for {
			line, err = con.ReadLine()
			if err != nil {
				return
			}
			if DEBUG {
				log.Println("<<", line)
			}
			inbound <- parse(line)
		}
	}()
	c = &client{con: con, textchan: outbound, inbound: inbound}
	go c.inbounds()
	go c.outbounds()
	return
}
