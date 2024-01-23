package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

type apiConfig struct {
	fileserverHits int
}

func main() {
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *dbg {
		deleteDB()
	}

	rChi := chi.NewRouter()
	rAPI := chi.NewRouter()
	rAdmin := chi.NewRouter()

	corsSrvMux := middlewareCors(rChi)

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

	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	rChi.Handle("/app/*", fsHandler)
	rChi.Handle("/app", fsHandler)

	rAPI.Get("/healthz", handleHealthz)
	rAPI.HandleFunc("/reset", apiCfg.handleReset)

	rAPI.Post("/chirps", handlePostChirps)
	rAPI.Get("/chirps", handleGetChirps)
	rAPI.Post("/users", handlePostUsers)

	rAPI.Get("/chirps/{chirpID}", handleChirpID)

	rAPI.Post("/login", handleLogin)

	rAdmin.Get("/metrics", apiCfg.handleMetrics)

	// mount namespaces routers to /api
	rChi.Mount("/api", rAPI)
	rChi.Mount("/admin", rAdmin)

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

func deleteDB() {
	err := os.Remove("internal/database/database.json")

	if err != nil {
		log.Fatal(err)
	}
}
