package gobus

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)


type Notify struct {
	Operation string
	ID int
}

const NotifyKey = "Notify"

type PubSub struct {
	Publish func(table string, operation string, id int) error
	Subscribe func(table string, fn func(operation string, id int) error) (unsubscribe func())
}

func NewPubSub() PubSub {
	bus := NewBus()

	return PubSub {
			Subscribe: func(table string, fn func(operation string, id int) error) (unsubscribe func()) {
				return bus.Subscribe(table, func (ctx context.Context) error {
					notify := ctx.Value(NotifyKey).(Notify)
					return fn(notify.Operation, notify.ID)
				})
			},
			Publish: func(table string, operation string, id int) error {
				return bus.Publish(table, context.WithValue(context.Background(), NotifyKey, Notify {
					Operation: operation,
					ID: id,
				}))
			},
	}
}

const TestKey = "test"

func TestRaw(t *testing.T) {
	bus := NewBus()
	var called int
	bus.Subscribe("hello", func(ctx context.Context) error {
		called = ctx.Value(TestKey).(int)
		return nil
	})
	require.NoError(t, bus.Publish("hello", context.WithValue(context.Background(), TestKey, 11)))
	require.Equal(t, 11, called)
}

func TestEmpty(t *testing.T) {
	bus := NewBus()
	require.NoError(t, bus.Publish("hello", context.Background()))
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
