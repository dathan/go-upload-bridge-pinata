package upload

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

//Build a structured response which allows a key value payload
type ResponseEnvelope struct {
	Status  string                 `json:"status"`
	Msg     string                 `json:"msg"`
	Payload map[string]interface{} `json:"payload"`
}

// build the structure correctly
func NewResponse() *ResponseEnvelope {

	resp := &ResponseEnvelope{
		Status:  "unknown",
		Msg:     "",
		Payload: make(map[string]interface{}),
	}

	return resp
}

//ReponseEnvelopers writes reponse to the response writer or error
func (s *ResponseEnvelope) WriteResponse(w http.ResponseWriter) error {
	jbyte, err := json.Marshal(s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(jbyte); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	log.Infof("STRUCTURED_RESPONSE: %s", string(jbyte))
	return nil
}
