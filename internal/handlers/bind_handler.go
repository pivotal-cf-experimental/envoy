package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/pivotal-cf-experimental/envoy/domain"
)

type Binder interface {
	Bind(domain.BindRequest) (domain.BindResponse, error)
}

type BindHandler struct {
	binder Binder
}

func NewBindHandler(binder Binder) BindHandler {
	return BindHandler{
		binder: binder,
	}
}

func (handler BindHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	request := handler.Parse(req)

	response, err := handler.binder.Bind(request)
	if err != nil {
		if err == domain.ServiceBindingAlreadyExistsError {
			respond(w, http.StatusConflict, EmptyJSON)
		} else {
			respond(w, http.StatusInternalServerError, Failure{
				Description: err.Error(),
			})
		}
		return
	}

	respond(w, http.StatusCreated, struct {
		Credentials    domain.BindingCredentials `json:"credentials,omitempty"`
		SyslogDrainURL string                    `json:"syslog_drain_url,omitempty"`
	}{
		Credentials:    response.Credentials,
		SyslogDrainURL: response.SyslogDrainURL,
	})
}

func (handler BindHandler) Parse(req *http.Request) domain.BindRequest {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	var params struct {
		ServiceID string `json:"service_id"`
		PlanID    string `json:"plan_id"`
		AppGUID   string `json:"app_guid"`
	}
	err = json.Unmarshal(body, &params)
	if err != nil {
		panic(err)
	}

	expression := regexp.MustCompile(`^/v2/service_instances/(.*)/service_bindings/(.*)$`)
	matches := expression.FindStringSubmatch(req.URL.Path)

	return domain.BindRequest{
		BindingID:  matches[2],
		InstanceID: matches[1],
		ServiceID:  params.ServiceID,
		PlanID:     params.PlanID,
		AppGUID:    params.AppGUID,
	}
}
