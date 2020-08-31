# gobus

Reflect-free event bus. Soon to be dated by generics. :)

This is just an experiment to see what a type safe event bus would look like.

The underlying bus is swaddled in closures by the caller to achieve an effect similar to sort.

This might be practical for an application that only needs one event bus, e.g: if you were dispatching based on pg/notify events.

Realistically, this is probably a bit too weird for general use. :D

```go
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
					// avoid nested locking problems by returning a unit of work
					return func () error {
						return fn(operation, id)
					}
				})
			},
	}
}

func main() {
    ps := NewPubSub()

    unsubscribe := ps.Subscribe("users", func (operation string, id int) error {
    	// ...
    })

    defer unsubscribe()

    ps.Publish("users", "UPDATE", 1)
}
```
