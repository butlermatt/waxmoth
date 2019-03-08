package server

import (
	"net/http"
)

/// Server is the local HTTP server for configuring and viewing local stations.
type Server struct {
	s *http.ServeMux
}

func New() *Server {
	s := &Server{s: http.NewServeMux()}
	s.s.HandleFunc("/", s.index)
	return s
}

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
