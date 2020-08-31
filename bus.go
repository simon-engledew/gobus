package gobus // import "github.com/simon-engledew/gobus"

import (
	"context"
	"fmt"
	"sync"

	"github.com/segmentio/ksuid"
)

type BusID ksuid.KSUID
type Handler func (ctx context.Context) error

type Bus struct {
	mutex     sync.RWMutex
	listeners map[string]map[BusID]Handler
	nextID    BusID
}

func NewBus() *Bus {
	return &Bus{
		listeners: make(map[string]map[BusID]Handler),
	}
}

func (bus *Bus) Subscribe(key string, fn Handler) (unsubscribe func()) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	eventID := bus.nextID

	bus.nextID = BusID(ksuid.New())

	indexes, found := bus.listeners[key]

	if !found {
		indexes = make(map[BusID]Handler)
		bus.listeners[key] = indexes
	}

	bus.listeners[key][eventID] = fn

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
		})
	}
}


func (bus *Bus) listenersFor(key string) []Handler {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()
	listeners, ok := bus.listeners[key]
	if !ok {
		return []Handler {}
	}
	output := make([]Handler, 0, len(listeners))
	for _, listener := range listeners {
		output = append(output, listener)
	}
	return output
}


func (bus *Bus) Publish(key string, ctx context.Context) error {
	for _, listener := range bus.listenersFor(key) {
		if err := listener(ctx); err != nil {
			return err
		}
	}
	return nil
}