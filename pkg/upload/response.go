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

func SetupCORS(w *http.ResponseWriter) {
	log.Infof("Sending CORS response")
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

//ReponseEnvelopers writes reponse to the response writer or error
func (s *ResponseEnvelope) WriteResponse(w http.ResponseWriter) {
	log.Info("Writing Response")
	jbyte, err := json.Marshal(s)
	if err != nil {
		log.Warnf("Error with JSON: %s", err)
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	SetupCORS(&w)
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(jbyte); err != nil {
		log.Warnf("Issue with response: %s", err)
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("STRUCTURED_RESPONSE: %s", string(jbyte))
	return
}

func Error(w http.ResponseWriter, err string, code int) {
	log.Warnf("Response Error: %s [%d]", err, code)
	SetupCORS(&w)
	http.Error(w, err, http.StatusInternalServerError)
}
