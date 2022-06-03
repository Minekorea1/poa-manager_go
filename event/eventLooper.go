package event

import (
	"container/list"
	"sync"
)

type Event struct {
	target EventTarget
	name   EventName
	args   []interface{}
}

type EventLooper struct {
	listeners map[EventTarget]func(EventName, []interface{})
	eventList *list.List

	mutex *sync.Mutex
	cond  *sync.Cond
}

func NewEventLooper() *EventLooper {
	mutex := &sync.Mutex{}
	cond := sync.NewCond(mutex)
	return &EventLooper{listeners: map[EventTarget]func(EventName, []interface{}){}, eventList: list.New(), mutex: mutex, cond: cond}
}

func (evtLooper *EventLooper) RegisterEventHandler(evtType EventTarget, handler func(EventName, []interface{})) {
	evtLooper.listeners[evtType] = handler
}

func (evtLooper *EventLooper) PushEvent(target EventTarget, name EventName, args ...interface{}) {
	evtLooper.eventList.PushBack(Event{target: target, name: name, args: args})

	evtLooper.cond.Signal()
}

func (evtLooper *EventLooper) Loop() {
	go func() {
		for {
			evtLooper.mutex.Lock()

			if evtLooper.eventList.Len() > 0 {
				element := evtLooper.eventList.Front()
				evtLooper.eventList.Remove(element)
				evtLooper.mutex.Unlock()

				event := element.Value.(Event)
				listener := evtLooper.listeners[event.target]
				if listener != nil {
					listener(event.name, event.args)
				}
			} else {
				evtLooper.cond.Wait()
				evtLooper.mutex.Unlock()
			}
		}
	}()
}
