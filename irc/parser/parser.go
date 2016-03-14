package parser

import (
	"log"
	"strconv"
	"strings"
)

var DEBUG = false

func debugf(f string, i ...interface{}) {
	log.Printf("PARSER: "+f, i...)
}

type Event interface {
	Text() string
}

type msg struct {
	text string
}

func (m *msg) Text() string {
	return m.text
}

type msgping struct {
	*msg
}

func (m *msgping) Token() (tok string) {
	parts := strings.Split(m.text, " ")
	tok = parts[1]
	return
}

type msgtarg struct {
	*msg
	source  string
	target  string
	payload string
}

func (m *msgtarg) Source() string {
	return m.source
}

func (m *msgtarg) Target() string {
	return m.target
}

func (m *msgtarg) Payload() string {
	return m.payload
}

type msgnotice struct {
	*msgtarg
	notice string
}

func (m *msgnotice) Notice() string {
	return m.notice
}

type msgpriv struct {
	*msgtarg
	priv string
}

type msgjoin struct {
	*msgtarg
	joined string
}

func (m *msgjoin) Joined() string {
	return m.joined
}

func (m *msgpriv) Priv() string {
	return m.priv
}

type msgnum struct {
	*msgtarg
	num int
}

func (m *msgnum) Num() int {
	return m.num
}

func Parse(line string) (e Event) {
	msg := &msg{line}
	parts := strings.Split(line, " ")
	switch {
	case strings.Index(line, "PING ") == 0:
		if DEBUG {
			debugf("PING %s", line)
		}
		return &msgping{msg}
	case len(parts) > 2:
		mt := &msgtarg{msg: msg, source: parts[0], target: parts[2]}
		if len(parts) > 3 {
			mt.payload = strings.Join(parts[3:], " ")
		}
		switch parts[1] {
		case "NOTICE":
			return &msgnotice{mt, mt.payload[1:]}
		case "PRIVMSG":
			return &msgpriv{mt, mt.payload[1:]}
		case "JOIN":
			return &msgjoin{mt, mt.target[1:]}
		default:
			if num, err := strconv.Atoi(parts[1]); err == nil {
				return &msgnum{mt, num}
			}
		}

	/*

	 */
	default:
	}
	return msg
}
