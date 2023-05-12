package health

import (
	"encoding/json"
	"net/http"
)

type liveness struct {
	Healthy bool
}

type readiness struct {
	Ready bool
}

func LivenessProbe(w http.ResponseWriter, r *http.Request) {
	ret := liveness{Healthy: true}

	resp, err := json.Marshal(ret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func ReadinessProbe(w http.ResponseWriter, r *http.Request) {
	ret := readiness{Ready: true}

	resp, err := json.Marshal(ret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}
