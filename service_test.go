package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

type CutterService struct {
	*Service
}

func NewCutterService(parent *Service) *CutterService {
	return &CutterService{
		Service: NewService(parent),
	}
}

var _ = Describe("Service", func() {
	var adderService *AdderService
	var cutterService *CutterService
	var arr []int
	var pushNum func(*Event) bool
	var event Event
	BeforeEach(func() {
		adderService = NewAdderService()
		cutterService = NewCutterService(adderService.Service)
		arr = make([]int, 0)
		pushNum = func(event *Event) bool {
			arr = append(arr, event.Payload.(int))
			return true
		}
		event = Event{
			Variant: "someEvent",
			Payload: 12,
		}
	})
	It("can add and trigger an event handler", func() {
		adderService.AddEventListener(event.Variant, pushNum)
		adderService.Dispatch(&event)
		Expect(arr).To(HaveLen(1))
		Expect(arr[0]).To(Equal(12))
	})
	It("propagates the event to the parent", func() {
		adderService.AddEventListener(event.Variant, pushNum)
		cutterService.Dispatch(&event)
		Expect(arr).To(HaveLen(1))
		Expect(arr[0]).To(Equal(12))
	})
	When("the event handler returns false", func() {
		BeforeEach(func() {
			pushNum = func(event *Event) bool {
				return false
			}
		})
		It("does not propagate the event to the parent", func() {
			adderService.AddEventListener(event.Variant, pushNum)
			cutterService.Dispatch(&event)
			Expect(arr).To(HaveLen(0))
		})
	})
})
