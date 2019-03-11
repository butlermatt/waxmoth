package server

import (
	"net/http"
)

// Server is the local HTTP server for configuring and viewing local stations.
type Server struct {
	s *http.ServeMux
}

// New creates a new instance of Server
func New() *Server {
	s := &Server{s: http.NewServeMux()}
	s.s.HandleFunc("/", s.index)
	s.s.HandleFunc("/stations", s.stations)
	return s
}

// ListenAndServe will start the Server listening on the specified port.
func (s *Server) ListenAndServe(port string) error {
	return http.ListenAndServe(port, s.s)
}

func (s *Server) index(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" {
		http.ServeFile(w, req, "www/index.html")
		return
	}

	addr := req.URL.Path[1:]
	http.ServeFile(w, req, "www/"+addr)
}

func (s *Server) stations(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/stations" {
		http.NotFound(w, req)
		return
	}
}
