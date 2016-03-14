package ircproto

import (
	"crypto/tls"
	"github.com/j7b/ircproto/irc"
	"github.com/j7b/ircproto/events"
	"net"
	"time"
)

var DEBUG = false

type Client interface {
	Loop() bool
	Event() events.Event
	Err() error
	Notice(target,msg string) error
	PrivMsg(target,msg string) error
	Join(channel string, key ...string) error
	// Part(channel string) error
}

type dialer struct {
	username string
	usetls   bool
	timeout  *time.Duration
	dialer   *net.Dialer
}

func (d *dialer) UseTLS(v bool) {
	d.usetls = v
}

func (d *dialer) DialTimeout(dur time.Duration) {
	d.timeout = &dur
}

func (d *dialer) Dial(n, addr string) (client Client, err error) {
	var con net.Conn
	if d.timeout != nil {
		d.dialer.Timeout = *d.timeout
	}
	switch d.usetls {
	case true:
		con, err = tls.DialWithDialer(d.dialer, n, addr, nil)
	case false:
		con, err = d.dialer.Dial(n, addr)
	}
	if err != nil {
		return
	}
	if DEBUG {
		irc.DEBUG = true
	}
	client, err = irc.NewClient(con, d.username)
	return
}

func NewDialer(username string) (d *dialer) {
	d = &dialer{username: username, dialer: &net.Dialer{}}
	return
}
