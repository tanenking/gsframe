package constants

import (
	"sync"

	"github.com/tanenking/gsframe/gsinf"
)

var eventItemPool = sync.Pool{
	New: func() any {
		return &EventItem{
			params: make([]interface{}, 0, 20),
		}
	},
}

type EventItem struct {
	params []interface{}
}
type EventCB struct {
	cb_id int
	cb    func(args ...interface{})
}

type EventManager struct {
	cb_id         int
	events_cb_map sync.Map //*eventCB
	event_chan    chan *EventItem
	trigger_chan  chan struct{}
}

func NewEventManager() gsinf.IEventManager {
	t := &EventManager{
		cb_id:         0,
		events_cb_map: sync.Map{},
		event_chan:    make(chan *EventItem, 1024),
		trigger_chan:  make(chan struct{}, 256),
	}
	return t
}

func (m *EventManager) AddEventListener(cb func(args ...interface{})) int {
	m.cb_id++
	evtcb := &EventCB{
		cb_id: m.cb_id,
		cb:    cb,
	}
	m.events_cb_map.Store(evtcb.cb_id, evtcb)

	return evtcb.cb_id
}
func (m *EventManager) RemoveEventLister(cb_id int) {
	m.events_cb_map.Delete(cb_id)
}
func (m *EventManager) DispatchEvent(params ...interface{}) {
	t := eventItemPool.Get().(*EventItem)
	t.params = t.params[:0]
	t.params = append(t.params, params...)

	m.event_chan <- t
}

func (m *EventManager) Update() {
	for {
		select {
		case <-ExitChannel:
			return
		case evt, ok := <-m.event_chan:
			if ok && evt != nil {
				m.callback(evt)
				eventItemPool.Put(evt)
			}
		default:
			return
		}
	}
}

func (m *EventManager) callback(evt *EventItem) {
	m.events_cb_map.Range(func(key, value any) bool {
		cb, ok := value.(*EventCB)
		if ok && cb != nil {
			cb.cb(evt.params...)
		}
		return true
	})
}
