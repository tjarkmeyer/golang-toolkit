package httpencoder

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

func EncodeJson(i interface{}, w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(i)
}

func EncodeSuccesful(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode("succesful operation")
}

func EncodeFailed(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
}

func EncodeFileResponse(w http.ResponseWriter, r *http.Request, filename string, data []byte) {
	http.ServeContent(w, r, filename, time.Now(), bytes.NewReader(data))
}
