package service

type EventVariant string

type EventI interface {
	GetVariant() EventVariant
	GetPayload() interface{}
}
type EventHandler func(event EventI) (willPropagate bool)
