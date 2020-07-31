package gobus

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type PubSub struct {
	Publish func(table string, operation string, id int) error
	Subscribe func(table string, fn func(operation string, id int) error) (unsubscribe func())
}

func NewPubSub() PubSub {
	bus := NewBus()

	subscriptions := make(map[BusID]func(operation string, id int) error, 0)

	return PubSub {
			Subscribe: func(table string, fn func(operation string, id int) error) (unsubscribe func()) {
				return bus.Subscribe(table, func(idx BusID) {
					subscriptions[idx] = fn
				})
			},
			Publish: func(table string, operation string, id int) error {
				return bus.Publish(table, func(idx BusID) error {
					return subscriptions[idx](operation, id)
				})
			},
	}
}

func TestPublishSubscribe(t *testing.T) {
	ps := NewPubSub()

	var called int

	ps.Subscribe("users", func(operation string, id int) error {
		called = id
		return nil
	})
	require.NoError(t, ps.Publish("users", "UPDATE", 42))

	require.Equal(t, 42, called)
}

func TestPublishUnsubscribe(t *testing.T) {
	ps := NewPubSub()

	var called bool

	unsubscribe := ps.Subscribe("users", func(operation string, id int) error {
		called = true
		return nil
	})

	unsubscribe()

	require.NoError(t, ps.Publish("users", "UPDATE", 1))

	require.False(t, called)
}

func TestPublishSubscribeMiss(t *testing.T) {
	ps := NewPubSub()

	var called bool

	ps.Subscribe("users", func(operation string, id int) error {
		called = true
		return nil
	})
	require.NoError(t, ps.Publish("sessions", "UPDATE", 42))

	require.False(t, called)
}

func TestPublishSubscribeMany(t *testing.T) {
	ps := NewPubSub()

	var calledA, calledB bool

	ps.Subscribe("users", func(operation string, id int) error {
		calledA = true
		return nil
	})
	ps.Subscribe("users", func(operation string, id int) error {
		calledB = true
		return nil
	})
	require.NoError(t, ps.Publish("sessions", "UPDATE", 42))

	require.False(t, calledA)
	require.False(t, calledB)
}

