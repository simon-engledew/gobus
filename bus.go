package gobus // import "github.com/simon-engledew/gobus"

import (
	"fmt"
	"sync"
	"github.com/segmentio/ksuid"
)

type Bus struct {
	mutex sync.RWMutex
	listeners map[string]map[ksuid.KSUID]struct{}
	nextID ksuid.KSUID
}

func NewBus() *Bus {
	return &Bus {
		listeners: make(map[string]map[ksuid.KSUID]struct{}),
		nextID: ksuid.New(),
	}
}

type Unsubscribe func ()

func (bus *Bus) Subscribe(key string, fn func (idx ksuid.KSUID)) Unsubscribe {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	eventID := bus.nextID

	bus.nextID = ksuid.New()

	indexes, found := bus.listeners[key]

	if !found {
		indexes = make(map[ksuid.KSUID]struct{})
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


func (bus *Bus) SubscribeOnce(key string, fn func (idx ksuid.KSUID)) Unsubscribe {
	var once sync.Once
	var unsubscribe Unsubscribe

	unsubscribe = bus.Subscribe(key, func (idx ksuid.KSUID) {
		once.Do(func() {
			unsubscribe()

			fn(idx)
		})
	})

	return unsubscribe
}

func (bus *Bus) Publish(key string, fn func (idx ksuid.KSUID) error) error {
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