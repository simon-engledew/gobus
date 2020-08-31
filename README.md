# gobus

Basic, reflect-free event bus that uses Context

Can either be used 'raw':

```go
const TestKey = "TEST"

bus := NewBus()

var called int

bus.Subscribe("hello", func(ctx context.Context) error {
    called = ctx.Value(TestKey).(int)
    return nil
})

bus.Publish("hello", context.WithValue(context.Background(), TestKey, 11))

// called == 11
```

Or you can build a type-safe wrapper for more compile-time guarantees: 

```go
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

func main() {
    ps := NewPubSub()

    unsubscribe := ps.Subscribe("users", func (operation string, id int) error {
    	// ...
    })

    defer unsubscribe()

    ps.Publish("users", "UPDATE", 1)
}
```
