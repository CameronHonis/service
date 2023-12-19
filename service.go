package service

// PROBLEMS TO SOLVE FOR:
// 1. Since services are typically singletons, I need a way to reset them to prevent test pollution
// 2. I need a way to stub/mock services for testing.
//		A. Sometimes mocking is sufficient if I only care if that method gets called. More importantly I need to assert
//			that the method gets called with the correct arguments.
//		B. Sometimes I need to stub a service to return a specific value. This is useful when I want to test a specific
//			branch of code that depends on the return value of a service.
//		C. Sometimes I need to stub a service to return a specific error. This is useful when I want to test a specific
//			branch of code that depends on the return value of a service.
// 3. I need to handle side effects.
// 4. I need to inject configuration.

// SOLUTIONS:
// 1. I can create a Reset method on each service that resets all of its fields to their zero values with the "reflect" package.
// 			Q. What do pointers get zero-valued to?
//			A. nil
//			Q. How do we handle circular dependencies? Should we?
//			A. Circular dependencies are bad, so no.
// 2. I can create a "Stubbed" struct that embeds the service in itself and overrides all the service's methods in order
//     to intercept calls.
//		Q. Can I automate overriding methods on a service?
//		A. Nothing native, only using preprocessor steps before compilation
//
//		A. The Stubbed struct needs to implement a way to interact with mocks. It also needs to house the data
//			that the mocked method is called with.
//		B & C. As an extension, the Stubbed struct should also implement a way to set the return value/error of a mock.
//		D. The Stubbed struct should also be flushed for each test.
//			Q. Does this mean we only need a single constructor?
// 3. Two ways of handling side effects comes to mind:
//		1. Using an event handler.
//			A. To keep a non-circular dependency chain with this event handler service,
//				the event handler service should not contain any dependencies. The event handler should allow services to
//				establish side effects (multiple) linked to a dispatch variant. Order should not matter.
//			B. To allow for control over event propagation, each service on its own should allow for event handling and event
//				propagation.
//		2. Using defined methods on the service.
//			A. To react to field changes, both of these methods require a setter method to intercept the change and
//				either dispatch an event or call a method directly.
//		Reasons to dispatch events over direct invocations:
//			* Easier to stub. I don't have to use a stub struct or define the entire interface to stub a single method
//			* Allows for dynamic event handler allocations
//			* Events to "bubble up" the chain of services, allows for side effects to be declared on other services
//				than the dispatcher service (better separation of concerns).
//		Reasons to use direct invocations over dispatching events:
//			* Faster since no cache miss (?)
//			* Easier to read. Can follow flow from "event" to side effect without understanding runtime setup
// 4. The trivial way of handling config would be to pass the config as an argument to the service constructor, but
//		this would not require config to be explicitly handled. Instead, I believe each service should explicitly declare
//		a config ingestor. Initial configs and all modified config objects should be injected into the app/root service
//		which then is passed down recursively and injected into the whole service tree.
//
//		To implement the injection on the service struct, it requires that the dependencies on the service are
//		stored internally on the service, rather than storing the dependencies on the service implementation.
//
//		It's not possible to call a per-service method implementation to be invoked from a generic service method.
//		instead of expecting each service to implement a config update handler, it should be the expectation that the
//		config does not change. This requires that the service builder takes a config to instantiate the service instance.
//		This also means that config injection should not be supported on the generic Service interface. Instead, each
//		service can implement their own custom config injection method.

type ServiceI interface {
	GetConfig() ConfigI
	AddDependency(service ServiceI)
	GetDependencies() []ServiceI
	Dispatch(event *Event)
	AddEventListener(eventVariant EventVariant, fn EventHandler) (eventId int)
	RemoveEventListener(eventId int)

	propagateEvent(event *Event)
}

type Service struct {
	parent                   ServiceI
	dependencies             []ServiceI
	config                   ConfigI
	eventHandlersCount       int
	variantByEventId         map[int]EventVariant
	eventHandlerIdxByEventId map[int]int
	eventHandlersByVariant   map[EventVariant][]EventHandler
}

func (s *Service) GetConfig() ConfigI {
	return s.config
}

func NewService(parent ServiceI) *Service {
	return &Service{
		parent:                   parent,
		config:                   nil,
		eventHandlersCount:       0,
		variantByEventId:         make(map[int]EventVariant),
		eventHandlerIdxByEventId: make(map[int]int),
		eventHandlersByVariant:   make(map[EventVariant][]EventHandler),
	}
}

func (s *Service) AddDependency(service ServiceI) {
	s.dependencies = append(s.dependencies, service)
}

func (s *Service) GetDependencies() []ServiceI {
	return s.dependencies
}

func (s *Service) Dispatch(event *Event) {
	willPropagate := true
	if eventHandlers, ok := s.eventHandlersByVariant[event.Variant]; ok {
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

func (s *Service) propagateEvent(event *Event) {
	if parentService, ok := s.parent.(*Service); ok {
		parentService.Dispatch(event)
	}
}
