package upload

//Build a structured response which allows a key value payload
type STRUCTURED_RESPONSE struct {
	Status  string                 `json:"status"`
	Msg     string                 `json:"msg"`
	Payload map[string]interface{} `json:"payload"`
}

func NewResponse() *STRUCTURED_RESPONSE {

	resp := &STRUCTURED_RESPONSE{
		Status:  "unknown",
		Msg:     "",
		Payload: make(map[string]interface{}),
	}

	return resp
}
