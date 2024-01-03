package service

import (
	"reflect"
)

type ServiceI interface {
	Config() ConfigI
	Dependencies() []ServiceI
	AddDependency(service ServiceI)
	Dispatch(event EventI)
	AddEventListener(eventVariant EventVariant, fn EventHandler) (eventId int)
	RemoveEventListener(eventId int)
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
		s.PropagateEvent(event)
	}
}

func (s *Service) Dependencies() []ServiceI {
	services := make([]ServiceI, 0)
	sVal := reflect.ValueOf(s.embeddedIn).Elem()
	fieldCount := sVal.NumField()
	for i := 0; i < fieldCount; i++ {
		fieldVal := sVal.Field(i)
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
	// validate dep
	depVal := reflect.ValueOf(dep).Elem()
	depType := depVal.Type()
	service := depVal.FieldByName("Service")
	if !service.IsValid() {
		panic("dependency does not embed Service")
	}

	// validate parent field exists
	expFieldName := depType.Name()
	parVal := reflect.ValueOf(s.embeddedIn).Elem()
	var parValField reflect.Value
	if parVal.FieldByName(expFieldName).IsValid() {
		parValField = parVal.FieldByName(expFieldName)
	} else {
		ServiceIType := reflect.TypeOf((*ServiceI)(nil)).Elem()
		depPtrType := reflect.ValueOf(dep).Type()
		for i := 0; i < parVal.NumField(); i++ {
			fieldVal := parVal.Field(i)
			fieldType := fieldVal.Type()
			if !fieldType.Implements(ServiceIType) {
				continue
			}
			if depPtrType.AssignableTo(fieldType) {
				parValField = fieldVal
				break
			}
		}
	}
	if !parValField.IsValid() {
		panic("could not determine the field on the parent for the dependency")
	}

	// set the parent on the dependency
	service.Addr().Interface().(*Service).SetParent(s.embeddedIn)

	// set the dependency on this service
	parValField.Set(reflect.ValueOf(dep))
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

func (s *Service) PropagateEvent(event EventI) {
	if s.parent == nil {
		return
	}
	s.parent.(ServiceI).Dispatch(event)
}

func (s *Service) SetParent(parent ServiceI) {
	s.parent = parent
}
