package pubsub

import (
	"sync"
)

const (
	// An internal/opaque name we can use for the SubscribeAll* functions.
	// The key doesn't matter as long as it's unlikely to collide with anything
	// anyone uses. (we could easily make it "").
	allName string = "__all__"
)

// Message wraps the custom event payload `Value` with top-level concepts
// used by the PubSub implementation.
type Message struct {
	Name  string
	Value interface{}
}

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

// SubscribeWithCallback provides a convenience helper around PubSub.Subscribe for
// a callback pattern.
// note: You still control the lifetime of your callback with the retruned channel instance
func SubscribeWithCallback(ps PubSub, name string, cb func(*Message)) chan<- interface{} {
	recv := make(chan *Message)
	unsub := ps.Subscribe(name, recv)

	go func() {
		for msg := range recv {
			cb(msg)
		}
	}()

	return unsub
}

// SubscribeAllWithCallback is like SubscribeWithCallback, but for all events
func SubscribeAllWithCallback(ps PubSub, cb func(*Message)) chan<- interface{} {
	return SubscribeWithCallback(ps, allName, cb)
}

// NewPubSub creates a new PubSub instance.
func NewPubSub() PubSub {
	return &pubsubImpl{
		receivers: make(map[string][]chan<- *Message),
	}
}

type pubsubImpl struct {
	lock      sync.RWMutex
	receivers map[string][]chan<- *Message
}

func (ps *pubsubImpl) Publish(msg *Message) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	// Inline message in the lock. This ensures the safety of the channel's lifetime
	// with being unsubscribed.
	for _, r := range ps.receivers[msg.Name] {
		r <- msg
	}

	for _, r := range ps.receivers[allName] {
		r <- msg
	}
}

func (ps *pubsubImpl) SubscribeAll(recv chan<- *Message) (closeChannel chan<- interface{}) {
	return ps.Subscribe(allName, recv)
}

func (ps *pubsubImpl) Subscribe(name string, recv chan<- *Message) chan<- interface{} {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	ps.receivers[name] = append(ps.receivers[name], recv)
	closeChannel := make(chan interface{})

	// Wait for subscriber to "unsubscribe"
	go func() {
		<-closeChannel

		// Always unblock the receiver
		defer close(recv)

		// Remove this receiver. This is not a hot path so naive rebuild is OK.
		ps.lock.Lock()
		defer ps.lock.Unlock()

		// safety/sanity check
		if len(ps.receivers[name]) != 0 {
			newRecievers := make([]chan<- *Message, 0, len(ps.receivers[name])-1)
			for _, r := range ps.receivers[name] {
				if r != recv {
					newRecievers = append(newRecievers, r)
				}
			}
			ps.receivers[name] = newRecievers
		}
	}()

	return closeChannel
}
