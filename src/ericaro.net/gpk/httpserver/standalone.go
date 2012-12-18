package httpserver

import (
	"log"
	"ericaro.net/gpk"
	"net/http"
	"time"
)

type StandaloneBackendServer struct {
	Local  gpk.LocalRepository // handles the real operations
	server http.Server
}

func (s *StandaloneBackendServer) Start(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/p/dl", func(w http.ResponseWriter, r *http.Request) { Send(s, w, r) })
	mux.HandleFunc("/p/ul", func(w http.ResponseWriter, r *http.Request) { Receive(s, w, r) })
	mux.HandleFunc("/p/nl", func(w http.ResponseWriter, r *http.Request) { Newer(s, w, r) })
	s.server = http.Server{
		Addr:    addr,
		Handler: mux,
	}
	s.server.ListenAndServe()
}

func (s *StandaloneBackendServer) Debugf(format string, args ...interface{}){
	log.Printf(format, args...)
}
	

//Contains return true if the server contains the ProjectID
func (s *StandaloneBackendServer) Receive(id gpk.ProjectID, timestamp time.Time, w http.ResponseWriter, r *http.Request) (err error) {
	_, err = s.Local.Install(r.Body)
	return
}

func (s *StandaloneBackendServer) Send(id gpk.ProjectID, w http.ResponseWriter, r *http.Request) {
	
	p, err := s.Local.FindPackage(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	p.Pack(w)
	return
}

func (s *StandaloneBackendServer) Newer(id gpk.ProjectID, timestamp time.Time, w http.ResponseWriter, r *http.Request) {
	p, err := s.Local.FindPackage(id)
	if err != nil || p == nil || !p.Timestamp().After(timestamp) {
		http.NotFound(w, r)
		return
	}
}