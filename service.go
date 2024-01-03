package service

import (
	"fmt"
	"reflect"
)

type ServiceI interface {
	Config() ConfigI
	Dependencies() []ServiceI
	AddDependency(service ServiceI)
	Dispatch(event EventI)
	AddEventListener(eventVariant EventVariant, fn EventHandler) (eventId int)
	RemoveEventListener(eventId int)

	propagateEvent(event EventI)
	setParent(parent ServiceI)
}

type Service struct {
	parent                   ServiceI
	embeddedIn               ServiceI
	config                   ConfigI
	eventHandlersCount       int
	variantByEventId         map[int]EventVariant
	eventHandlerIdxByEventId map[int]int
	eventHandlersByVariant   map[EventVariant][]EventHandler
}

func (s *Service) Config() ConfigI {
	return s.config
}

func NewService(service ServiceI, config ConfigI) *Service {
	return &Service{
		embeddedIn:               service,
		parent:                   nil,
		config:                   config,
		eventHandlersCount:       0,
		variantByEventId:         make(map[int]EventVariant),
		eventHandlerIdxByEventId: make(map[int]int),
		eventHandlersByVariant:   make(map[EventVariant][]EventHandler),
	}
}

func (s *Service) Dispatch(event EventI) {
	willPropagate := true
	if eventHandlers, ok := s.eventHandlersByVariant[event.Variant()]; ok {
		for _, eventHandler := range eventHandlers {
			if eventHandler == nil {
				continue
			}
			willPropagate = willPropagate && eventHandler(event)
		}
	}
	if eventHandlers, ok := s.eventHandlersByVariant[ALL_EVENTS]; ok {
		for _, eventHandler := range eventHandlers {
			if eventHandler == nil {
				continue
			}
			willPropagate = willPropagate && eventHandler(event)
		}
	}
	if willPropagate {
		s.propagateEvent(event)
	}
}

func (s *Service) Dependencies() []ServiceI {
	services := make([]ServiceI, 0)
	sVal := reflect.ValueOf(s.embeddedIn).Elem()
	sType := sVal.Type()
	fieldCount := sVal.NumField()
	for i := 0; i < fieldCount; i++ {
		fieldName := sType.Field(i).Name
		fieldVal := sVal.Field(i)
		fieldType := fieldVal.Type()
		// if I called this method from a struct that embeds Service, how would I get the embedding struct?
		fmt.Printf("\n%s (%s): %s\n", fieldName, fieldType, fieldVal)
		if !fieldVal.CanInterface() {
			continue
		}

		a := fieldVal.Interface()
		fieldService, ok := a.(ServiceI)
		if ok {
			services = append(services, fieldService)
		}
	}
	return services
}

func (s *Service) AddDependency(dep ServiceI) {
	// set the dependency on this service
	fieldName := reflect.TypeOf(dep).Elem().Name()
	parVal := reflect.ValueOf(s.embeddedIn).Elem()
	parVal.FieldByName(fieldName).Set(reflect.ValueOf(dep))

	// set the parent on the dependency
	dep.setParent(s.embeddedIn)
}

func (s *Service) AddEventListener(eventVariant EventVariant, fn EventHandler) (eventId int) {
	eventId = s.eventHandlersCount
	s.eventHandlersCount++
	if _, ok := s.eventHandlersByVariant[eventVariant]; !ok {
		s.eventHandlersByVariant[eventVariant] = make([]EventHandler, 0)
	}
	eventHandlerIdx := len(s.eventHandlersByVariant[eventVariant])
	s.variantByEventId[eventId] = eventVariant
	s.eventHandlerIdxByEventId[eventId] = eventHandlerIdx
	s.eventHandlersByVariant[eventVariant] = append(s.eventHandlersByVariant[eventVariant], fn)
	return eventId
}

func (s *Service) RemoveEventListener(eventId int) {
	variant, ok := s.variantByEventId[eventId]
	if !ok {
		return
	}
	eventHandlerIdx, ok := s.eventHandlerIdxByEventId[eventId]
	if !ok {
		return
	}
	s.eventHandlersByVariant[variant][eventHandlerIdx] = nil
	delete(s.variantByEventId, eventId)
	delete(s.eventHandlerIdxByEventId, eventId)
}

func (s *Service) propagateEvent(event EventI) {
	if s.parent == nil {
		return
	}
	s.parent.(ServiceI).Dispatch(event)
}

func (s *Service) setParent(parent ServiceI) {
	s.parent = parent
}
