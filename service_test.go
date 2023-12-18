package service

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type AdderService struct {
	*Service
	CutterService   *CutterService
	OtherSubService *OtherSubService
	fieldA          int
}

func NewAdderService() *AdderService {
	adderService := &AdderService{
		Service: NewService(nil),
		fieldA:  1,
	}
	adderService.CutterService = NewCutterService(adderService.Service)
	adderService.OtherSubService = NewSubservice(adderService.Service)
	return adderService
}

type CutterService struct {
	*Service
}

func NewCutterService(parent *Service) *CutterService {
	return &CutterService{
		Service: NewService(parent),
	}
}

type OtherSubService struct {
	*Service
	subserviceFieldA string
}

func NewSubservice(parent *Service) *OtherSubService {
	return &OtherSubService{
		Service:          NewService(parent),
		subserviceFieldA: "hello",
	}
}

type ServiceWithPrivateService struct {
	*Service
	privateFieldA  int
	privateService *OtherSubService
}

func NewServiceWithPrivateService() *ServiceWithPrivateService {
	rtn := &ServiceWithPrivateService{
		Service:       NewService(nil),
		privateFieldA: 1,
	}
	rtn.privateService = NewSubservice(rtn.Service)
	return rtn
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
	It("can remove an event handler", func() {

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

var _ = Describe("GetSubServices", func() {
	var adderService *AdderService
	BeforeEach(func() {
		adderService = NewAdderService()
	})
	It("retrieves subservices", func() {
		subservices := GetSubServices(adderService)
		Expect(subservices).To(HaveLen(2))
		Expect(subservices[0]).To(Equal(adderService.CutterService))
		Expect(subservices[1]).To(Equal(adderService.OtherSubService))
	})
	When("a private field exists", func() {
		var serviceWithPrivateService *ServiceWithPrivateService
		BeforeEach(func() {
			serviceWithPrivateService = NewServiceWithPrivateService()
		})
		It("does not evaluate the private field", func() {
			_ = GetSubServices(serviceWithPrivateService)
		})
	})
})
