package server

import (
	"encoding/json"
	"fmt"
	api "mem-db/pkg/api"
	router "mem-db/pkg/api/http/router"
	"net/http"
	"strings"
)

type WordResponse struct {
	Word        string `json:"word"`
	Occurrences int    `json:"occurrences"`
}

type TextInput struct {
	Text string `json:"text"`
}

type Response struct {
	Status     string         `json:"status"`
	StatusCode int            `json:"statusCode"`
	Data       []WordResponse `json:"data,omitempty"`
	Message    string         `json:"message,omitempty"`
}

type HTTPServer struct {
	router *router.Router
	server *http.Server
}

func NewServer(port int) api.Server {
	r := router.NewRouter()

	r.AddRoute("GET", "/words/occurences", getWordOccurences)
	r.AddRoute("POST", "/words/register", registerWords)

	return &HTTPServer{
		router: r,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: r,
		},
	}
}

func (s *HTTPServer) Start() error {

	fmt.Printf("Listening on %s.. \n", s.server.Addr)
	return s.server.ListenAndServe()
}

// GET /words/occurences?terms=apple,banana,orange
func getWordOccurences(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Query())
	query := r.URL.Query()
	terms := query["terms"]

	var results []WordResponse

	if len(terms) > 0 {

		terms[0] = strings.ToLower(terms[0])
		words := strings.Split(terms[0], ",")

		for _, word := range words {

			results = append(results, WordResponse{Word: word, Occurrences: 0})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(results)
	json.NewEncoder(w).Encode(&Response{
		Status:     "Success",
		StatusCode: http.StatusOK,
		Data:       results})

}

func registerWords(w http.ResponseWriter, r *http.Request) {
	var textInput *TextInput = &TextInput{}
	w.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(r.Body).Decode(textInput)
	if err != nil {
		json.NewEncoder(w).Encode(&Response{
			Status:     "Bad Request",
			StatusCode: http.StatusBadRequest,
			Message:    err.Error()})
		return

	}

	if len(textInput.Text) == 0 {
		json.NewEncoder(w).Encode(&Response{
			Status:     "Bad Request",
			StatusCode: http.StatusBadRequest,
			Message:    "Text field is empty"})
		return
	}

	// json.NewEncoder(w).Encode(fmt.Sprintf("Text processed successfully"))
	json.NewEncoder(w).Encode(&Response{
		Status:     "Success",
		StatusCode: http.StatusOK,
		Message:    "Text processed successfully"})
}
