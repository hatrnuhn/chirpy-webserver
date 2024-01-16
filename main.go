package main

import (
	"fmt"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
}

func main() {
	srvMux := http.NewServeMux()

	corsSrvMux := middlewareCors(srvMux)

	s := &http.Server{
		Addr:    "localhost:8080",
		Handler: corsSrvMux,
	}

	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	// If the handle path was "/app" instead of "/app/",
	// the server would only respond to exactly "/app", not
	// any subpaths like "/app/test.txt". Adding the
	// trailing slash allows the handler to match all paths
	// that start with "/app/", including "/app/" itself.

	// Then, by using http.StripPrefix, you're telling the
	// file server to ignore the "/app" prefix in the URL
	// request and look for the file directly under the
	// http.Dir() directory.

	srvMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	srvMux.HandleFunc("/metrics", apiCfg.handleMetrics)

	srvMux.HandleFunc("/healthz", handleHealthz)
	fmt.Printf("Starting server at http://%v\n", s.Addr)
	s.ListenAndServe()
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
