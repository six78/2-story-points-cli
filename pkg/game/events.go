package game

type EventTag int

const (
	EventStateChanged EventTag = iota
	EventAutoRevealScheduled
	EventAutoRevealCancelled
)

type Event struct {
	Tag  EventTag
	Data interface{}
}

type Subscription struct {
	Events chan Event
}

type EventPublisher interface {
	Publish(tag EventTag, data interface{})
}

type EventSubscriber interface {
	Subscribe(tag EventTag) *Subscription
}

type EventManager struct {
	subscriptions []*Subscription
}

func NewEventManager() *EventManager {
	return &EventManager{
		subscriptions: make([]*Subscription, 0, 1),
	}
}

func (m *EventManager) Send(event Event) {
	for _, sub := range m.subscriptions {
		sub.Events <- event
	}
}

func (m *EventManager) Subscribe() *Subscription {
	subscription := &Subscription{
		Events: make(chan Event, 10),
	}
	m.subscriptions = append(m.subscriptions, subscription)
	return subscription
}

func (m *EventManager) Count() int {
	return len(m.subscriptions)
}

func (m *EventManager) Close() {
	for _, sub := range m.subscriptions {
		close(sub.Events)
	}
	m.subscriptions = nil
}
