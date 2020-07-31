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

func (bus *Bus) Subscribe(key string, fn func(idx BusID)) (unsubscribe func()) {
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

	fn(eventID)

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


func (bus *Bus) Publish(key string, fn func(idx BusID) error) error {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()
	if listeners, ok := bus.listeners[key]; ok {
		for idx := range listeners {
			if err := fn(idx); err != nil {
				return err
			}
		}
	}
	return nil
}
