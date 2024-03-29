package service

type EventVariant string

const ALL_EVENTS EventVariant = "ALL_EVENTS"

type EventI interface {
	Variant() EventVariant
	Payload() interface{}
}

type Event struct {
	variant EventVariant
	payload interface{}
}

func NewEvent(variant EventVariant, payload interface{}) *Event {
	return &Event{
		variant: variant,
		payload: payload,
	}
}

func (e *Event) Variant() EventVariant {
	return e.variant
}

func (e *Event) Payload() interface{} {
	return e.payload
}

type EventHandler func(self ServiceI, event EventI) (willPropagate bool)
