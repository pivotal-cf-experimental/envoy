package handlers

import (
	"encoding/json"
	"net/http"
)

type Failure struct {
	Description string `json:"description"`
}

var EmptyJSON = map[string]interface{}{}

func respond(w http.ResponseWriter, code int, response interface{}) {
	body, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(body)
}
