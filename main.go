package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/hatrnuhn/chirpy-webserver/internal/database"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	fileserverHits int
	jwtSecret      string
	db             *database.DB
	polka          map[string]any
}

func main() {
	godotenv.Load()
	jwtSecret := os.Getenv("JWT_SECRET")

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
		jwtSecret:      jwtSecret,
		polka:          make(map[string]any),
	}
	apiCfg.polka["polkakey"] = os.Getenv("POLKA_KEY")

	var err error
	apiCfg.db, err = database.NewDB(os.Getenv("DBPATH"))
	if err != nil {
		log.Fatal("couldn't initialize database")
	}

	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	rChi.Handle("/app/*", fsHandler)
	rChi.Handle("/app", fsHandler)

	rAPI.Get("/healthz", handleHealthz)

	rAPI.HandleFunc("/reset", apiCfg.handleReset)

	rAPI.Post("/chirps", apiCfg.handlePostChirps)
	rAPI.Get("/chirps", apiCfg.handleGetChirps)
	rAPI.Post("/users", apiCfg.handlePostUsers)
	rAPI.Put("/users", apiCfg.handlePutUsers)
	rAPI.Get("/chirps/{chirpID}", apiCfg.handleChirpID)
	rAPI.Delete("/chirps/{chirpID}", apiCfg.handleDelChirpID)
	rAPI.Post("/login", apiCfg.handlePostLogin)
	rAPI.Post("/refresh", apiCfg.handlePostRefresh)
	rAPI.Post("/revoke", apiCfg.handlePostRevoke)
	rAPI.Post("/polka/webhooks", apiCfg.handlePostPolkaWebhooks)

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
