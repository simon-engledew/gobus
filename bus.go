package gobus // import "github.com/simon-engledew/gobus"

import (
	"fmt"
	"sync"

	"github.com/segmentio/ksuid"
)

type BusID ksuid.KSUID

type Bus struct {
	mutex     sync.RWMutex
	listeners map[string]map[BusID]struct{}
	nextID    BusID
}

func NewBus() *Bus {
	return &Bus{
		listeners: make(map[string]map[BusID]struct{}),
	}
}

func (bus *Bus) Subscribe(key string, onSubscribe func(idx BusID), onUnsubscribe func (idx BusID)) (unsubscribe func()) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	eventID := bus.nextID

	bus.nextID = BusID(ksuid.New())

	indexes, found := bus.listeners[key]

	if !found {
		indexes = make(map[BusID]struct{})
		bus.listeners[key] = indexes
	}

	bus.listeners[key][eventID] = struct{}{}

	onSubscribe(eventID)

	var once sync.Once

	return func() {
		once.Do(func() {
			bus.mutex.Lock()
			defer bus.mutex.Unlock()

			if interfaces, found := bus.listeners[key]; found {
				if _, ok := interfaces[eventID]; !ok {
					panic(fmt.Errorf("subscription missing: %s %v", key, eventID))
				}
				delete(interfaces, eventID)
				if len(interfaces) == 0 {
					delete(bus.listeners, key)
				}
			} else {
				panic(fmt.Errorf("no subscriptions for %s found", key))
			}

			onUnsubscribe(eventID)
		})
	}
}


func (bus *Bus) handlers(key string, fn func(idx BusID) func () error) []func () error {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()
	if listeners, ok := bus.listeners[key]; ok {
		output := make([]func () error, 0, len(listeners))
		for idx := range listeners {
			output = append(output, fn(idx))
		}
		return output
	}
	return []func () error {}
}


func (bus *Bus) Publish(key string, fn func(idx BusID) func () error) error {
	for _, handler := range bus.handlers(key, fn) {
		if err := handler(); err != nil {
			return err
		}
	}
	return nil
}