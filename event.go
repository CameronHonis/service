package service

type EventVariant string

type Event struct {
	Variant EventVariant
	Payload interface{}
}
type EventHandler func(event *Event) (willPropagate bool)
