# gobus

Simple, reflect-free event bus. Soon to be dated by generics. :)

```go
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

func main() {
    publish, subscribe := NewPubSub()

    defer subscribe("users", func (operation string, id int) error {
    	// ...
    })()

    publish("users", "UPDATE", 1)
}
```