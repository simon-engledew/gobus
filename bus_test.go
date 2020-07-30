package gobus

import (
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
	"testing"
)


type DatabaseSubscribe func (table string, fn func (operation string, id int) error) Unsubscribe
type DatabasePublish func (table string, operation string, id int) error

func NewPubSub() (DatabasePublish, DatabaseSubscribe) {
	bus := NewBus()

	subscriptions := make(map[ksuid.KSUID]func (operation string, id int) error, 0)

	subscribe := func (table string, fn func (operation string, id int) error) Unsubscribe {
		return bus.Subscribe(table, func (idx ksuid.KSUID) {
			subscriptions[idx] = fn
		})
	}

	publish := func (table string, operation string, id int) error {
		return bus.Publish(table, func (idx ksuid.KSUID) error {
			return subscriptions[idx](operation, id)
		})
	}

	return publish, subscribe
}

func TestPublishSubscribe(t *testing.T) {
	publish, subscribe := NewPubSub()

	var called int

	subscribe("users", func (operation string, id int) error {
		called = id
		return nil
	})
	require.NoError(t, publish("users", "UPDATE", 42))

	require.Equal(t, 42, called)
}

func TestPublishUnsubscribe(t *testing.T) {
	publish, subscribe := NewPubSub()

	var called bool

	unsubscribe := subscribe("users", func (operation string, id int) error {
		called = true
		return nil
	})

	unsubscribe()

	require.NoError(t, publish("users","UPDATE", 1))

	require.False(t, called)
}


func TestPublishSubscribeMiss(t *testing.T) {
	publish, subscribe := NewPubSub()

	var called bool

	subscribe("users", func (operation string, id int) error {
		called = true
		return nil
	})
	require.NoError(t, publish("sessions", "UPDATE", 42))

	require.False(t, called)
}