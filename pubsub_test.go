package pubsub

import (
	"fmt"
	"sync"
	"testing"
)

func TestSequence(t *testing.T) {
	var recvWg sync.WaitGroup
	var unsubWg sync.WaitGroup

	ps := NewPubSub()

	recvPickles := make(chan *Message)
	unsubPickles := ps.Subscribe("pickles", recvPickles)

	// This is our receiver
	recvWg.Add(1)
	unsubWg.Add(1)
	go func() {
		// Loop until unsubscribed
		for msg := range recvPickles {
			fmt.Printf("recv'd message %+v\n", msg)

			if msg.Name != "pickles" {
				t.Errorf("Wrong name %s", msg.Name)
			}
			v, ok := msg.Value.(int)
			if !ok || v != 99 {
				t.Errorf("Wrong value %+v", msg.Value)
			}
			recvWg.Done()
		}

		// Indicate unsubscribe completed
		fmt.Println("done")
		unsubWg.Done()
	}()

	// This is some decoupled publisher
	go func() {
		msg := &Message{
			Name:  "pickles",
			Value: 99,
		}
		fmt.Printf("Publishing %+v\n", msg)
		ps.Publish(msg)
	}()

	// Make sure the reciever got his message
	recvWg.Wait()

	// Can tear down the receiver now.
	close(unsubPickles)
	unsubWg.Wait()
}

func TestAllWithCallback(t *testing.T) {
	var recvWg sync.WaitGroup

	ps := NewPubSub()

	// This is our receiver
	recvWg.Add(1)
	unsubPickles := SubscribeAllWithCallback(ps, func(msg *Message) {
		fmt.Printf("recv'd message %+v\n", msg)

		if msg.Name != "pickles" {
			t.Fatalf("Wrong name %s", msg.Name)
		}
		v, ok := msg.Value.(int)
		if !ok || v != 99 {
			t.Errorf("Wrong value %+v", msg.Value)
		}
		recvWg.Done()
	})

	// This is some decoupled publisher
	go func() {
		msg := &Message{
			Name:  "pickles",
			Value: 99,
		}
		fmt.Printf("Publishing %+v\n", msg)
		ps.Publish(msg)
	}()

	// Make sure the reciever got his message
	recvWg.Wait()
	close(unsubPickles)
}
