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
type PubSub struct {
	Publish       func(table string, operation string, id int) error
	Subscribe     func(table string, fn func(operation string, id int) error) (unsubscribe func())
	SubscribeOnce func(table string, fn func(operation string, id int) error) (unsubscribe func())
}

type ctxKeyNotify int

const notifyKey ctxKeyNotify = 0

type notify struct {
	operation string
	id        int
}

func NewPubSub() *PubSub {
	bus := NewBus()

	return &PubSub{
		Subscribe: func(table string, fn func(operation string, id int) error) (unsubscribe func()) {
			return bus.Subscribe(table, func(ctx context.Context) error {
				notify := ctx.Value(notifyKey).(notify)
				return fn(notify.operation, notify.id)
			})
		},
		SubscribeOnce: func(table string, fn func(operation string, id int) error) (unsubscribe func()) {
			return bus.SubscribeOnce(table, func(ctx context.Context) error {
				notify := ctx.Value(notifyKey).(notify)
				return fn(notify.operation, notify.id)
			})
		},
		Publish: func(table string, operation string, id int) error {
			return bus.Publish(table, context.WithValue(context.Background(), notifyKey, notify{
				operation: operation,
				id:        id,
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
