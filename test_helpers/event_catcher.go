package test_helpers

import (
	"fmt"
	"github.com/CameronHonis/marker"
	. "github.com/CameronHonis/service"
	"sync"
)

type EventCatcher struct {
	Service
	__Dependencies__ marker.Marker
	ListeningTo      ServiceI

	__State__    marker.Marker
	evs          []EventI
	evsByVariant map[EventVariant][]EventI
	mu           sync.Mutex
}

func NewEventCatcher() *EventCatcher {
	ec := &EventCatcher{
		evs:          make([]EventI, 0),
		evsByVariant: make(map[EventVariant][]EventI),
	}
	ec.Service = *NewService(ec, nil)
	catcher := func(_ ServiceI, ev EventI) bool {
		ec.CatchEvent(ev)
		return false
	}
	ec.AddEventListener(ALL_EVENTS, catcher)
	return ec
}

func (ec *EventCatcher) FlushEvents() {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.evs = make([]EventI, 0)
	ec.evsByVariant = make(map[EventVariant][]EventI)
}

func (ec *EventCatcher) CatchEvent(ev EventI) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.evs = append(ec.evs, ev)
	if _, ok := ec.evsByVariant[ev.Variant()]; !ok {
		ec.evsByVariant[ev.Variant()] = make([]EventI, 0)
	}
	ec.evsByVariant[ev.Variant()] = append(ec.evsByVariant[ev.Variant()], ev)
}

func (ec *EventCatcher) LastEvent() EventI {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	if len(ec.evs) == 0 {
		panic("no events have been caught")
	}
	return ec.evs[len(ec.evs)-1]
}

func (ec *EventCatcher) LastEventByVariant(eVar EventVariant) EventI {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	evs, ok := ec.evsByVariant[eVar]
	if !ok {
		panic(fmt.Sprintf("no events with variant %s have been caught", eVar))
	}
	return evs[len(evs)-1]
}

func (ec *EventCatcher) EventsCount() int {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	return len(ec.evs)
}

func (ec *EventCatcher) EventsByVariantCount(eVar EventVariant) int {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	evs, ok := ec.evsByVariant[eVar]
	if !ok {
		return 0
	}
	return len(evs)
}

func (ec *EventCatcher) NthEvent(idx int) EventI {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	if idx >= len(ec.evs) {
		panic(fmt.Sprintf("idx %d exceeds bounds of caught events (size %d)", idx, len(ec.evs)))
	}
	return ec.evs[idx]
}

func (ec *EventCatcher) NthEventByVariant(eVar EventVariant, idx int) EventI {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	evs, ok := ec.evsByVariant[eVar]
	if !ok {
		panic(fmt.Sprintf("no %s events have been caught", eVar))
	}
	if idx >= len(evs) {
		panic(fmt.Sprintf("idx %d exceeds bounds of caught %s events (size %d)", idx, eVar, len(evs)))
	}
	return evs[idx]
}
