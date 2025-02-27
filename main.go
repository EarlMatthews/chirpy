package main

import (
	"net/http"
)



func main(){
	mux := http.NewServeMux()

	// Create a New server
	srv := http.Server{
		Addr: ":8888",
		Handler: mux,
	}

	// Use the http.NewServeMux's .Handle() method to add a handler for the root path (/).
	mux.Handle("/", http.FileServer(http.Dir(".")))
	//Starting the new server
	if err := srv.ListenAndServe(); err != nil {
        panic(err)
    }
}