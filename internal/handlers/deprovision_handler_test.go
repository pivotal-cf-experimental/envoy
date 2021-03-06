package handlers_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/pivotal-cf-experimental/envoy/domain"
	"github.com/pivotal-cf-experimental/envoy/internal/handlers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Deprovisioner struct {
	WasCalledWith    domain.DeprovisionRequest
	WasCalled        bool
	DeprovisionError error
}

func (d *Deprovisioner) Deprovision(deprovisionRequest domain.DeprovisionRequest) error {
	d.WasCalledWith = deprovisionRequest
	d.WasCalled = true
	return d.DeprovisionError
}

func NewDeprovisioner() *Deprovisioner {
	return &Deprovisioner{}
}

var _ = Describe("DeprovisionHandler", func() {
	var deprovisioner *Deprovisioner
	var handler handlers.DeprovisionHandler

	BeforeEach(func() {
		deprovisioner = NewDeprovisioner()
		handler = handlers.NewDeprovisionHandler(deprovisioner)
	})

	It("calls the deprovisioner Deprovision method with the correct values", func() {
		writer := httptest.NewRecorder()

		url := fmt.Sprintf("%s?plan_id=%s&service_id=%s",
			"/v2/service_instances/service-instance-id",
			"the-1gb-plan-id",
			"the-sshfs-service-id")
		request, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			panic(err)
		}

		handler.ServeHTTP(writer, request)

		Expect(deprovisioner.WasCalledWith).To(Equal(domain.DeprovisionRequest{
			InstanceID: "service-instance-id",
			ServiceID:  "the-sshfs-service-id",
			PlanID:     "the-1gb-plan-id",
		}))

	})

	Context("when the deprovisioner succeeds", func() {
		It("returns a 200 OK with JSON {}", func() {
			writer := httptest.NewRecorder()
			request, err := http.NewRequest("DELETE",
				"/v2/service_instances/service-instance-id?plan_id=some-plan-id&service_id=some-service-id",
				nil)
			if err != nil {
				panic(err)
			}

			handler.ServeHTTP(writer, request)

			Expect(writer.Code).To(Equal(http.StatusOK))
			Expect(writer.Header()["Content-Type"]).To(Equal([]string{"application/json"}))

			Expect(writer.Body.String()).To(MatchJSON("{}"))

		})
	})

	Context("when the service instance does not exist", func() {
		It("returns a 410 Gone with JSON {}", func() {
			writer := httptest.NewRecorder()
			request, err := http.NewRequest("DELETE",
				"/v2/service_instances/a-missing-service-instance-id?plan_id=some-plan-id&service_id=some-service-id",
				nil)
			if err != nil {
				panic(err)
			}

			deprovisioner.DeprovisionError = domain.ServiceInstanceNotFoundError("that instance doesn't exist!")

			handler.ServeHTTP(writer, request)

			Expect(writer.Code).To(Equal(http.StatusGone))
			Expect(writer.Header()["Content-Type"]).To(Equal([]string{"application/json"}))
			Expect(writer.Body.String()).To(MatchJSON("{}"))
		})
	})

	Context("when the deprovisioner fails", func() {
		It("returns a 500 error with the message", func() {
			writer := httptest.NewRecorder()
			request, err := http.NewRequest("DELETE", "/v2/service_instances/service-instance-id?plan_id=some-plan-id&service_id=some-service-id",
				nil)
			if err != nil {
				panic(err)
			}

			deprovisioner.DeprovisionError = errors.New("my database failed somehow!")

			handler.ServeHTTP(writer, request)

			Expect(writer.Code).To(Equal(http.StatusInternalServerError))
			Expect(writer.Header()["Content-Type"]).To(Equal([]string{"application/json"}))
			Expect(writer.Body.String()).To(MatchJSON(`{"description": "my database failed somehow!"}`))
		})
	})

	Context("when the request is missing a required parameter", func() {
		It("should not call the deprovisioner", func() {
			writer := httptest.NewRecorder()

			url := "/v2/service_instances/service-instance-id?plan_id=some-plan-id"
			request, err := http.NewRequest("DELETE", url, nil)
			if err != nil {
				panic(err)
			}

			handler.ServeHTTP(writer, request)

			Expect(deprovisioner.WasCalled).To(BeFalse())
		})

		It("should return a 400 error with a helpful message", func() {
			writer := httptest.NewRecorder()

			url := "/v2/service_instances/service-instance-id?plan_id=some-plan-id"
			request, err := http.NewRequest("DELETE", url, nil)
			if err != nil {
				panic(err)
			}

			handler.ServeHTTP(writer, request)

			Expect(writer.Code).To(Equal(http.StatusBadRequest))
			Expect(writer.Header()["Content-Type"]).To(Equal([]string{"application/json"}))
			var msg struct {
				Description string `json:"description"`
			}
			Expect(json.Unmarshal(writer.Body.Bytes(), &msg)).To(Succeed())
			Expect(msg.Description).To(ContainSubstring("service_id"))
		})
	})
})
