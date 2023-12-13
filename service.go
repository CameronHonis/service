package main

import (
	"fmt"
	"reflect"
)

type Action struct {
	variant string
	payload interface{}
}

type ServiceI interface {
	Reset()
	Dispatch(action Action)
	AddListener(actionVariant string, listener func(payload interface{}))
	RemoveListener(actionVariant string) error
}

type Service struct {
}

func Reset(s ServiceI) {
	sVal := reflect.ValueOf(s).Elem()
	sType := sVal.Type()
	fieldCount := sVal.NumField()
	for i := 0; i < fieldCount; i++ {
		field := sType.Field(i)
		fieldVal := sVal.Field(i)
		fmt.Println("field name: ", field.Name)
		fmt.Println("field type: ", field.Type)
		fmt.Println("field val: ", fieldVal)
	}
}

type AdderService struct {
	Service
	fieldA int
	FieldB bool
}

func NewAdderService() *AdderService {
	return &AdderService{
		fieldA: 0,
		FieldB: false,
	}
}
func (as *AdderService) Reset() {
	as.fieldA = 0
	as.FieldB = false
}
func (as *AdderService) Dispatch(action Action) {

}
func (as *AdderService) AddListener(actionVariant string, listener func(payload interface{})) {

}
func (as *AdderService) RemoveListener(actionVariant string) error {
	return nil
}
func main() {
	adderService := NewAdderService()
	Reset(adderService)
}
