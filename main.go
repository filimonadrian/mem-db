package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Router struct {
	routes map[string]map[string]http.HandlerFunc
}

type WordResponse struct {
	Word        string `json:"word"`
	Occurrences int    `json:"occurrences"`
}

type TextInput struct {
	Text string `json:"text"`
}

type Config struct {
	Port int `json:"port"`
}

type Response struct {
	Status     string         `json:"status"`
	StatusCode int            `json:"statusCode"`
	Data       []WordResponse `json:"data,omitempty"`
	Message    string         `json:"message,omitempty"`
}

func NewRouter() *Router {
	router := &Router{
		routes: make(map[string]map[string]http.HandlerFunc),
	}

	router.addRoute("GET", "/words/occurences", getWordOccurences)
	router.addRoute("POST", "/words/register", registerWords)

	return router
}

// add CORS headers to the response
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET")
		w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Authorization, Origin, application/json")

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func (r *Router) addRoute(method, path string, handlerFunc http.HandlerFunc) {
	if r.routes[path] == nil {
		r.routes[path] = make(map[string]http.HandlerFunc)
	}
	r.routes[path][method] = handlerFunc
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if handlers, ok := r.routes[req.URL.Path]; ok {
			if handlerFunc, methodExists := handlers[req.Method]; methodExists {
				handlerFunc(w, req)
				return
			}
		}
		http.NotFound(w, req)
	}))

	handler.ServeHTTP(w, req)
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

func StartServer(config Config) {

	router := NewRouter()

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: router,
	}

	fmt.Printf("listening to port %d.. \n", config.Port)
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func main() {

	config := Config{
		Port: 8080,
	}

	StartServer(config)
}
