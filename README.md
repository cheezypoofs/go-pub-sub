# go-pub-sub
A simple in-process pub/sub for golang

# Motivation

Call it somewhere between "I spent no more than 5 minutes looking for one that existed" and "go is such a pleasure to use, any excuse to code something is a good one".

Mostly, what I wanted is a way to use the benefits of the pub/sub model in process to decouple unrelated components. I wanted an excuse to use channels to "share by communicating" and because channels are more idiomatic than callbacks (though I did still implement a callback helper).

# Godoc

```
package pubsub // import "github.com/cheezypoofs/go-pub-sub"


FUNCTIONS

func SubscribeAllWithCallback(ctx context.Context, ps PubSub, cb func(Message))
    SubscribeAllWithCallback is like SubscribeWithCallback, but for all events

func SubscribeWithCallback(ctx context.Context, ps PubSub, name string, cb func(Message))
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
	Publish(msg Message)

	// Subscribe creates a subscription on the topic `name`. You are responsible
	// for watching the returned channel until the context is canceled.
	Subscribe(ctx context.Context, name string) <-chan Message

	// SubscribeAll is like Subscribe, but you will receive all events.
	SubscribeAll(ctx context.Context) <-chan Message
}

func NewPubSub() PubSub
    NewPubSub creates a new PubSub instance.
```
