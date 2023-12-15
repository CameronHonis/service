package main

import (
	. "github.com/onsi/ginkgo/v2"
)

type AdderService struct {
	*Service
	fieldA int
}

func NewAdderService() *AdderService {
	return &AdderService{
		Service: NewService(nil),
		fieldA:  2,
	}
}

var _ = Describe("Service", func() {
	var adderService *AdderService
	BeforeEach(func() {
		adderService = NewAdderService()
	})
	It("can add and trigger an event handler", func() {
		arr := make([]int, 0)
		pushTwelve := func(event *Event) {

		}
		adderService.AddEventListener()
	})
})
