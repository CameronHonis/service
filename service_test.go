package service

import (
	. "github.com/CameronHonis/marker"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type AdderConfig struct {
	ConfigFieldOne int
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

type AdderService struct {
	Service

	__dependencies__           Marker
	CutterService              *CutterService
	OtherSubService            *OtherSubService
	StrangelyNamedServiceField *StrangeService

	__state__ Marker
	fieldA    int
}

func NewAdderService(config *AdderConfig) *AdderService {
	adderService := &AdderService{
		fieldA: 1,
	}
	adderService.Service = *NewService(adderService, config)
	return adderService
}

type CutterService struct {
	Service

	__dependencies__ Marker

	__state__ Marker
}

func NewCutterService(config *CutterConfig) *CutterService {
	cutterService := &CutterService{}
	cutterService.Service = *NewService(cutterService, config)
	return cutterService
}

type OtherSubService struct {
	Service
	subserviceFieldA string
}

type StrangeService struct {
	Service
}

func NewStrangeService() *StrangeService {
	s := &StrangeService{}
	s.Service = *NewService(s, nil)
	return s
}

func NewSubService(config *OtherSubServiceConfig) *OtherSubService {
	otherSubService := &OtherSubService{
		subserviceFieldA: "hello",
	}
	otherSubService.Service = *NewService(otherSubService, config)
	return otherSubService
}

func BuildServices() *AdderService {
	adderService := NewAdderService(&AdderConfig{})
	cutterService := NewCutterService(&CutterConfig{})
	otherSubService := NewSubService(&OtherSubServiceConfig{})
	strangeService := NewStrangeService()

	adderService.AddDependency(cutterService)
	adderService.AddDependency(otherSubService)
	adderService.AddDependency(strangeService)

	return adderService
}

var _ = Describe("Service", func() {
	var adderService *AdderService
	var arr []int
	var pushNum func(EventI) bool
	var event Event
	BeforeEach(func() {
		adderService = BuildServices()
		arr = make([]int, 0)
		pushNum = func(event EventI) bool {
			arr = append(arr, event.Payload().(int))
			return true
		}
		event = Event{
			variant: "someEvent",
			payload: 12,
		}
	})
	It("can add and trigger an event handler", func() {
		adderService.AddEventListener(event.Variant(), pushNum)
		adderService.Dispatch(&event)
		Expect(arr).To(HaveLen(1))
		Expect(arr[0]).To(Equal(12))
	})
	It("can add and trigger a universal event handler", func() {
		adderService.AddEventListener(ALL_EVENTS, pushNum)
		adderService.Dispatch(&event)
		Expect(arr).To(HaveLen(1))
		Expect(arr[0]).To(Equal(12))
	})
	It("propagates the event to the parent", func() {
		adderService.AddEventListener(event.Variant(), pushNum)
		adderService.CutterService.Dispatch(&event)
		Expect(arr).To(HaveLen(1))
		Expect(arr[0]).To(Equal(12))
	})
	When("an existing handler is removed", func() {
		var handlerId int
		BeforeEach(func() {
			handlerId = adderService.AddEventListener(event.Variant(), pushNum)
		})
		It("does not fire that event handler", func() {
			adderService.RemoveEventListener(handlerId)
			adderService.Dispatch(&event)
			Expect(arr).To(HaveLen(0))
		})
	})
	When("the event handler returns false", func() {
		BeforeEach(func() {
			pushNum = func(event EventI) bool {
				return false
			}
		})
		It("does not propagate the event to the parent", func() {
			adderService.AddEventListener(event.Variant(), pushNum)
			adderService.CutterService.Dispatch(&event)
			Expect(arr).To(HaveLen(0))
		})
	})
})

var _ = Describe("Dependencies", func() {
	var adderService *AdderService
	BeforeEach(func() {
		adderService = BuildServices()
	})
	It("retrieves subservices", func() {
		deps := adderService.Dependencies()
		Expect(deps).To(HaveLen(3))
	})
})

var _ = Describe("AddDependency", func() {
	var adderService *AdderService
	var cutterService *CutterService
	type NonsenseService struct {
		Service
	}
	BeforeEach(func() {
		adderService = NewAdderService(&AdderConfig{})
		cutterService = NewCutterService(&CutterConfig{})
	})

	It("adds the dependency as a field on adderService", func() {
		adderService.AddDependency(cutterService)
		Expect(adderService.CutterService).To(Equal(cutterService))
	})
	It("sets the parent of the dependency to adderService", func() {
		adderService.AddDependency(cutterService)
		Expect(cutterService.parent).To(Equal(adderService))
	})
	When("the parent service does not have a field for the dependency", func() {
		It("panics", func() {
			Expect(func() {
				adderService.AddDependency(&NonsenseService{})
			}).To(Panic())
		})
	})
	When("the parent service does have a field for the dependency, but is named ambiguously", func() {
		var strangeService *StrangeService
		BeforeEach(func() {
			strangeService = NewStrangeService()
		})
		It("sets the dependency on the ambiguously named field", func() {
			adderService.AddDependency(strangeService)
			Expect(strangeService.parent).To(Equal(adderService))
		})
	})
})
