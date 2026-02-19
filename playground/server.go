package playground

import (
	"log"
	"net/http"
	"os"
)

// Server handles playground requests
type Server struct{}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()

	// Serve static files
	dir := "assets/playground"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Printf("Warning: Static directory %s not found. Current dir: %v", dir, func() string { d, _ := os.Getwd(); return d }())
	} else {
		log.Printf("Serving static files from %s", dir)
	}

	fs := http.FileServer(http.Dir(dir))
	
	// Handle / returning index.html
	mux.Handle("/", fs)

	log.Printf("Starting playground server at http://%s", addr)
	return http.ListenAndServe(addr, mux)
}
