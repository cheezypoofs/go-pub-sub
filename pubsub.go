package pubsub

import (
	"context"
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
	Publish(msg Message)

	// Subscribe creates a subscription on the topic `name`. You are responsible
	// for watching the returned channel until the context is canceled.
	Subscribe(ctx context.Context, name string) <-chan Message

	// SubscribeAll is like Subscribe, but you will receive all events.
	SubscribeAll(ctx context.Context) <-chan Message
}

// SubscribeWithCallback provides a convenience helper around PubSub.Subscribe for
// a callback pattern.
// note: You still control the lifetime of your callback with the retruned channel instance
func SubscribeWithCallback(ctx context.Context, ps PubSub, name string, cb func(Message)) {
	go func() {
		ch := ps.Subscribe(ctx, name)
		for msg := range ch {
			cb(msg)
		}
	}()
}

// SubscribeAllWithCallback is like SubscribeWithCallback, but for all events
func SubscribeAllWithCallback(ctx context.Context, ps PubSub, cb func(Message)) {
	SubscribeWithCallback(ctx, ps, allName, cb)
}

// NewPubSub creates a new PubSub instance.
func NewPubSub() PubSub {
	return &pubSubImpl{
		subscriptions: make(map[string][]*subscription),
	}
}

type subscription struct {
	ch chan Message
}

type pubSubImpl struct {
	lock          sync.RWMutex
	subscriptions map[string][]*subscription
}

func (ps *pubSubImpl) Publish(msg Message) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	// Inline message in the lock. This ensures the safety of the channel's lifetime
	// with being unsubscribed.
	for _, s := range ps.subscriptions[msg.Name] {
		s.ch <- msg
	}

	for _, s := range ps.subscriptions[allName] {
		s.ch <- msg
	}
}

func (ps *pubSubImpl) Subscribe(ctx context.Context, name string) <-chan Message {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	sub := &subscription{
		ch: make(chan Message),
	}
	ps.subscriptions[name] = append(ps.subscriptions[name], sub)

	go func() {
		// Block until receiver says they are done.
		<-ctx.Done()

		// Mark we are done
		close(sub.ch)

		// Remove this receiver. This is not a hot path so naive rebuild is OK.
		ps.lock.Lock()
		defer ps.lock.Unlock()

		// safety/sanity check
		if len(ps.subscriptions[name]) != 0 {
			newSubs := make([]*subscription, 0, len(ps.subscriptions[name])-1)
			for _, s := range ps.subscriptions[name] {
				if s != sub {
					newSubs = append(newSubs, s)
				}
			}
			ps.subscriptions[name] = newSubs
		}
	}()

	return sub.ch
}

func (ps *pubSubImpl) SubscribeAll(ctx context.Context) <-chan Message {
	return ps.Subscribe(ctx, allName)
}
