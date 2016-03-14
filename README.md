# ircproto
IRC Client Protocol Implementation

## Sequence

```
d := NewDialer("ircusername")
client,err := d.Dial("tcp","irc.server.example:6666")
check(err)
for client.Loop() {
	e := client.Event()
	dosomethingwith(e)
}
```
