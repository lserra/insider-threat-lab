package detection

import (
	"encoding/json"
	"net/http"
)

func Handler(report Report) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(response http.ResponseWriter, request *http.Request) {
		response.WriteHeader(http.StatusOK)
		_, _ = response.Write([]byte("ok\n"))
	})
	mux.HandleFunc("/metrics", func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = response.Write([]byte(Prometheus(report)))
	})
	mux.HandleFunc("/report", func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(response).Encode(report); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
	})
	return mux
}
