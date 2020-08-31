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
				}, func (idx BusID) {
					delete(subscriptions, idx)
				})
			},
			Publish: func(table string, operation string, id int) error {
				return bus.Publish(table, func(idx BusID) func () error {
					fn := subscriptions[idx]
					return func () error {
						return fn(operation, id)
					}
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
	require.NoError(t, ps.Publish("users", "UPDATE", 42))

	require.True(t, calledA)
	require.True(t, calledB)
}


func TestPublishNested(t *testing.T) {
	ps := NewPubSub()

	var calledA, calledB int

	ps.Subscribe("users", func(operation string, id int) error {
		calledA = id
		return ps.Publish("sessions", "UPDATE", 1)
	})
	ps.Subscribe("sessions", func(operation string, id int) error {
		calledB = id
		return nil
	})
	require.NoError(t, ps.Publish("users", "UPDATE", 42))

	require.Equal(t, 42, calledA)
	require.Equal(t, 1, calledB)
}


func TestSubscribeNested(t *testing.T) {
	ps := NewPubSub()

	var calledA, calledB int

	ps.Subscribe("users", func(operation string, id int) error {
		calledA = id
		ps.Subscribe("sessions", func(operation string, id int) error {
			calledB = id
			return nil
		})
		return nil
	})

	require.NoError(t, ps.Publish("users", "UPDATE", 42))
	require.NoError(t, ps.Publish("sessions", "UPDATE", 1))

	require.Equal(t, 42, calledA)
	require.Equal(t, 1, calledB)
}
