package main

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

// SOLUTIONS:
// 1. I can create a Reset method on each service that resets all of its fields to their zero values with the "reflect" package.
// 			Q. What do pointers get zero-valued to?
//			A. nil
//			Q. How do we handle circular dependencies? Should we?
//			A. Circular dependencies are bad
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
//		A. Using an event handler. To keep a non-circular dependency chain with this event handler service,
//			the event handler service should not contain any dependencies. The event handler should allow services to
//			establish side effects (multiple) linked to a dispatch variant. Order should not matter.
//		B. To allow for control over event propagation, each service on its own should allow for event handling and event
//			propagation.

type EventVariant string

type Event struct {
	Variant EventVariant
	Payload interface{}
}
type EventHandler func(interface{})

type ServiceI interface {
	Dispatch(event *Event)
	AddEventListener(eventVariant EventVariant, fn EventHandler) (eventId int)
	RemoveEventListener(eventId int)

	propagateEvent(event *Event)
}

type Service struct {
	parent                ServiceI
	eventHandlersById     map[int]EventHandler
	eventHandlerByVariant map[EventVariant]EventHandler
}

func NewService(parent ServiceI) *Service {
	return &Service{
		parent,
		make(map[int]EventHandler),
		make(map[EventVariant]EventHandler),
	}
}

func (s *Service) Dispatch(event *Event) {
	if eventHandler, ok := s.eventHandlerByVariant[event.Variant]; ok {
		eventHandler(event)
	}
}

func (s *Service) AddEventListener(eventVariant EventVariant, fn EventHandler) (eventId int) {
	return 0
}

func (s *Service) RemoveEventListener(eventId int) {

}

func (s *Service) propagateEvent(event *Event) {

}

func main() {

}
