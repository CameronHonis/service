package service

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type AdderService struct {
	*Service
	config *AdderService
	fieldA int
}

func NewAdderService() *AdderService {
	adderService := &AdderService{
		Service: NewService(nil),
		fieldA:  1,
	}
	adderService.AddDependency(NewCutterService(adderService.Service))
	adderService.AddDependency(NewSubservice(adderService.Service))
	return adderService
}

func (as *AdderService) ingestConfig(config ConfigI) {
}

type CutterService struct {
	*Service
	config *CutterConfig
}

func NewCutterService(parent *Service) *CutterService {
	return &CutterService{
		Service: NewService(parent),
	}
}

func (cs *CutterService) ingestConfig(config ConfigI) {
}

type OtherSubService struct {
	*Service
	config           *OtherSubServiceConfig
	subserviceFieldA string
}

func NewSubservice(parent *Service) *OtherSubService {
	return &OtherSubService{
		Service:          NewService(parent),
		subserviceFieldA: "hello",
	}
}

func (cs *OtherSubService) ingestConfig(config ConfigI) {
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
	When("an existing handler is removed", func() {
		var handlerId int
		BeforeEach(func() {
			handlerId = adderService.AddEventListener(event.Variant, pushNum)
		})
		It("does not fire that event handler", func() {
			adderService.RemoveEventListener(handlerId)
			adderService.Dispatch(&event)
			Expect(arr).To(HaveLen(0))
		})
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

var _ = Describe("GetDependencies", func() {
	var adderService *AdderService
	BeforeEach(func() {
		adderService = NewAdderService()
	})
	It("retrieves subservices", func() {
		deps := adderService.GetDependencies()
		Expect(deps).To(HaveLen(2))
	})
})

type AdderConfig struct {
	ConfigFieldOne        int
	CutterConfig          *CutterConfig
	OtherSubServiceConfig *OtherSubServiceConfig
}

func (ac *AdderConfig) MergeWith(config ConfigI) ConfigI {
	configCopy := *(config.(*AdderConfig))
	return &configCopy
}

type CutterConfig struct {
	IsHotBlade bool
}

func (ac *CutterConfig) MergeWith(config ConfigI) ConfigI {
	configCopy := *(config.(*CutterConfig))
	return &configCopy
}

type OtherSubServiceConfig struct {
	OtherSubServiceSecret string
}

func (subServiceConfig *OtherSubServiceConfig) MergeWith(config ConfigI) ConfigI {
	configCopy := *(config.(*OtherSubServiceConfig))
	return &configCopy
}
