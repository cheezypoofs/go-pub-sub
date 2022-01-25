# go-pub-sub
A simple in-process pub/sub for golang

# Motivation

Call it somewhere between "I spent no more than 5 minutes looking for one that existed" and "go is such a pleasure to use, any excuse to code something is a good one".

Mostly, what I wanted is a way to use the benefits of the pub/sub model in process to decouple unrelated components. I wanted an excuse to use channels to "share by communicating" and because channels are more idiomatic than callbacks (though I did still implement a callback helper).

# Godoc

```
FUNCTIONS

func SubscribeAllWithCallback(ps PubSub, cb func(*Message)) chan<- interface{}
    SubscribeAllWithCallback is like SubscribeWithCallback, but for all events

func SubscribeWithCallback(ps PubSub, name string, cb func(*Message)) chan<- interface{}
    SubscribeWithCallback provides a convenience helper around PubSub.Subscribe
    for a callback pattern. note: You still control the lifetime of your
    callback with the retruned channel instance


TYPES

type Message struct {
	Name  string
	Value interface{}
}
    Message wraps the custom event payload `Value` with top-level concepts used
    by the PubSub implementation.

type PubSub interface {
	// Publish the message to all subscribers of msg.Name
	Publish(msg *Message)

	// Subscribe creates a subscription on the topic `name`. You provide the
	// channel on which you will listen for Message's and can loop on that waiting for it
	// to close. You unsubscribe by closing the returned channel. This will in turn
	// close your `recv` channel so that your listener can exit.
	// note: You should receive and process as quickly as possible as the entire
	// PubSub instance is blocked on your channel brokering the message.
	Subscribe(name string, recv chan<- *Message) (closeChannel chan<- interface{})

	// SubscribeAll is like Subscribe, but you will receive all events.
	SubscribeAll(recv chan<- *Message) (closeChannel chan<- interface{})
}

func NewPubSub() PubSub
    NewPubSub creates a new PubSub instance.
```
