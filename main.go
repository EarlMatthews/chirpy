package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct{
	fileserverHits atomic.Int32
}

type Chirp struct {
	Body string `json:"body"`
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not POST", http.StatusMethodNotAllowed)
		return
	}

	var chirp Chirp
	err := json.NewDecoder(r.Body).Decode(&chirp)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Here you can add your validation logic for the chirp struct
	if len(chirp.Body) > 140 {
		http.Error(w, "{\"error\":\"Chirp is too long\"}", http.StatusBadRequest)
		return
	}

	// If validation passes, you can respond with success or any other appropriate response
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	response := map[string]bool{"valid": true}
	json.NewEncoder(w).Encode(response)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	// Next, write a new middleware method on a *apiConfig that increments the fileserverHits counter every time it's called
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increment the fileserverHits counter
		cfg.fileserverHits.Add(1)
		
		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

func healthz(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type","text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}

func metrics (cfg *apiConfig, w http.ResponseWriter, r *http.Request){
	// Create a new handler that writes the number of requests 
	// that have been counted as plain text in this format to the HTTP response
	count := cfg.fileserverHits.Load()
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type","text/plain; charset=utf-8")
	html := fmt.Sprintf(`
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>`, count)
	_, _ = w.Write([]byte(html))
}

func reset (cfg *apiConfig, w http.ResponseWriter, r *http.Request){
	//Finally, create and register a handler on the /reset path that,
	//when hit, will reset your fileserverHits back to 0
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type","text/plain; charset=utf-8")
}

func main(){
	
	mux := http.NewServeMux()
	cfg := &apiConfig{fileserverHits: atomic.Int32{}}
	// Create a New server
	srv := http.Server{
		Addr: ":8888",
		Handler: mux,
	}

	// Use the http.NewServeMux's .Handle() method to add a handler for the root path (/).
	
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.Handle("/app/assets/", cfg.middlewareMetricsInc(http.StripPrefix("/app/assets/", http.FileServer(http.Dir("assets")))))
	mux.HandleFunc("GET /admin/healthz",healthz)
	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics(cfg, w, r)
	} )
	mux.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, r *http.Request) {
		reset(cfg, w, r)
	} )
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)

	//Starting the new server
	if err := srv.ListenAndServe(); err != nil {
        panic(err)
    }
}
