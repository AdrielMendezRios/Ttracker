package api

import "net/http"

// TODO: Implement API server
func StartServer() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // TODO: Add proper request handling
    })
}
