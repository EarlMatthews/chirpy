package main

import (
	"net/http"
)

func healthz(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type","text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}

func main(){
	mux := http.NewServeMux()

	// Create a New server
	srv := http.Server{
		Addr: ":8888",
		Handler: mux,
	}

	// Use the http.NewServeMux's .Handle() method to add a handler for the root path (/).
	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("."))))
	//mux.Handle("/assets/",http.FileServer(http.Dir("assets")))
	mux.Handle("/app/assets/", http.StripPrefix("/app/assets/", http.FileServer(http.Dir("assets"))))
	mux.HandleFunc("/healthz",healthz)

	//Starting the new server
	if err := srv.ListenAndServe(); err != nil {
        panic(err)
    }
}
