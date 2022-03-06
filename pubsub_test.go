package pubsub

import (
	"context"
	"sync"
	"testing"
)

func TestSequence(t *testing.T) {
	ps := NewPubSub()

	var recvWg sync.WaitGroup
	var unsubWg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())

	// This is our receiver
	recvWg.Add(1)
	unsubWg.Add(1)
	go func() {
		for msg := range ps.Subscribe(ctx, "pickles") {
			t.Logf("recv'd message %+v\n", msg)

			if msg.Name != "pickles" {
				t.Errorf("Wrong name %s", msg.Name)
			}
			v, ok := msg.Value.(int)
			if !ok || v != 99 {
				t.Errorf("Wrong value %+v", msg.Value)
			}
			recvWg.Done()
		}

		t.Logf("done")
		unsubWg.Done()
	}()

	// This is some decoupled publisher
	go func() {
		msg := Message{
			Name:  "pickles",
			Value: 99,
		}
		t.Logf("Publishing %+v\n", msg)
		ps.Publish(msg)
	}()

	// Make sure the reciever got his message
	recvWg.Wait()

	// Can tear down the receiver now.
	cancel()
	unsubWg.Wait()
}

func TestAllWithCallback(t *testing.T) {

	var recvWg sync.WaitGroup
	ps := NewPubSub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// This is our receiver
	recvWg.Add(1)
	SubscribeAllWithCallback(ctx, ps, func(msg Message) {
		t.Logf("recv'd message %+v\n", msg)

		if msg.Name != "pickles" {
			t.Errorf("Wrong name %s", msg.Name)
		}
		v, ok := msg.Value.(int)
		if !ok || v != 99 {
			t.Errorf("Wrong value %+v", msg.Value)
		}
		recvWg.Done()
	})

	// This is some decoupled publisher
	go func() {
		msg := Message{
			Name:  "pickles",
			Value: 99,
		}
		t.Logf("Publishing %+v\n", msg)
		ps.Publish(msg)
	}()

	// Make sure the reciever got his message
	recvWg.Wait()
}
