package events

import (

)

type Event interface {
	Text() string
}

type Ping interface {
	Event
	Token() string
}

type Targeted interface {
	Event
	Source() string
	Target() string
	Payload() string
}

type Join interface {
	Joined() string
}

type Notice interface {
	Targeted
	Notice() string
}

type Priv interface {
	Targeted
	Priv() string
}

type Numeric interface {
	Targeted
	Num() int
}
