# gobus

Simple, reflect-free event bus. Soon to be dated by generics. :)

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
				return bus.Publish(table, func(idx BusID) error {
					return subscriptions[idx](operation, id)
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
