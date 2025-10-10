package main

import (
	"Confeet/internal/manager"
	"fmt"
	"log"
	"net/http"
)

func main() {
	mgr := manager.NewManager()

	http.Handle("/", http.FileServer(http.Dir("/frontend")))
	http.HandleFunc("/ws", mgr.ServeWS)

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
