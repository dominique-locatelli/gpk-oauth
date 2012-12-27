//protocol defines the basic remote protocols between a client and a remote repository
// it defines a Client interface, and an http implementation. There is also a directory based implementation available in the localrepository package
package protocol

import (
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strconv"
)

const (
	FETCH  = "fetch"
	PUSH   = "push"
	SEARCH = "search"
)

type ProtocolError struct {
	Message string
	Code int
}
func (p *ProtocolError) Error() string {return p.Message}


var (
	StatusForbidden = &ProtocolError{"Forbidden Operation", http.StatusForbidden}
	StatusIdentityMismatch = &ProtocolError{"Mismatch between Identity Declared and Received", http.StatusExpectationFailed}
	StatusCannotOverwrite = &ProtocolError{"Cannot Overwrite a Package", http.StatusConflict}
	StatusMissingDependency = &ProtocolError{"Missing Dependency", http.StatusPartialContent}
)

func ErrorCode( err error ) int {
	switch e:= err.(type) {
		case *ProtocolError:
			return e.Code
	}
	return http.StatusInternalServerError
}


type Server interface {
	Receive(pid PID, r io.ReadCloser) (error)
	Serve(pid PID, w io.Writer) (error)
	Search(query string, start int) ([]PID, error)
	Debugf(format string, args ...interface{})
}

func Handle(p string, s Server) { HandleMux(p, s, http.DefaultServeMux) }

func HandleMux(p string, s Server, mux *http.ServeMux) {
	mux.HandleFunc(path.Join(p, PUSH), func(w http.ResponseWriter, r *http.Request) {
		servePush(s, w, r)
	})
	mux.HandleFunc(path.Join(p, FETCH), func(w http.ResponseWriter, r *http.Request) {
		serveFetch(s, w, r)
	})
	mux.HandleFunc(path.Join(p, SEARCH), func(w http.ResponseWriter, r *http.Request) {
		serveSearch(s, w, r)
	})
}

//Receive HandlerFunc that s
func servePush(s Server, w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" { // on the push URL only POST method are supported
		http.Error(w, "Method not supported.", http.StatusMethodNotAllowed)
		return
	}

	// identify the package
	vals := r.URL.Query()
	pid, err := FromParameter(&vals)
	if err != nil {
		s.Debugf("Error %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = s.Receive(*pid, r.Body) // create and fill the blob
	if err != nil {
		http.Error(w, err.Error(), ErrorCode(err))
	}
	// can pass the reason as body response)
	//	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	//	fmt.Fprintln(w, error)

}

func serveFetch(s Server, w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	pid, err := FromParameter(&vals)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if pid.Name == "" {
		http.NotFound(w, r)
		return
	}

	err = s.Serve(*pid, w)
	if err != nil {
		http.Error(w, err.Error(), ErrorCode(err) )
	}
	return
}

func serveSearch(s Server, w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("q")
	start, _ := strconv.Atoi(r.FormValue("start"))
	results, err := s.Search(query, start)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

	} else {
		json.NewEncoder(w).Encode(results)
	}
}
