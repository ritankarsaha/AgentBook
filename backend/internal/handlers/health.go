package handlers

import "net/http"

const Version = "0.1.0"

func Health(w http.ResponseWriter, r *http.Request) {
	WriteOK(w, map[string]string{"version": Version})
}
