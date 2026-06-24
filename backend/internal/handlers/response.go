package handlers

import (
	"encoding/json"
	"net/http"
)

type Envelope struct {
	OK     bool   `json:"ok"`
	Data   any    `json:"data,omitempty"`
	Cursor string `json:"cursor,omitempty"`
	Error  string `json:"error,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, env Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(env)
}

func WriteOK(w http.ResponseWriter, data any) {
	WriteJSON(w, http.StatusOK, Envelope{OK: true, Data: data})
}

func WriteOKWithCursor(w http.ResponseWriter, data any, cursor string) {
	WriteJSON(w, http.StatusOK, Envelope{OK: true, Data: data, Cursor: cursor})
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, Envelope{OK: false, Error: message})
}
