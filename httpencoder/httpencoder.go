package httpencoder

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type IHttpEncoder interface {
	EncodeJson(interface{}, http.ResponseWriter, int)
	EncodeSuccesful(http.ResponseWriter, int)
	EncodeFailed(http.ResponseWriter, int)
	EncodeFileResponse(http.ResponseWriter, *http.Request, string, []byte)
}

type HttpEncoder struct{}

func New() *HttpEncoder {
	return &HttpEncoder{}
}

func (h *HttpEncoder) EncodeJson(i interface{}, w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(i)
}

func (h *HttpEncoder) EncodeSuccesful(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode("succesful operation")
}

func (h *HttpEncoder) EncodeFailed(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
}

func (h *HttpEncoder) EncodeFileResponse(w http.ResponseWriter, r *http.Request, filename string, data []byte) {
	http.ServeContent(w, r, filename, time.Now(), bytes.NewReader(data))
}
