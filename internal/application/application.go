package application

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/YattaDeSune/calc-project/pkg/calculation"
)

type Config struct {
	Addr string
}

func ConfigFromEnv() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{Addr: port}
}

type Application struct {
	config *Config
}

func New() *Application {
	return &Application{config: ConfigFromEnv()}
}

type Request struct {
	Expression string `json:"expression"`
}

type Response struct {
	Result float64 `json:"result,omitempty"`
	Error  string  `json:"error,omitempty"`
}

func CalcHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Иннициализация декодера для десириализации
	decoder := json.NewDecoder(r.Body)
	var request Request
	err := decoder.Decode(&request)

	// Иннициализация энкодера для сериализации
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)

	// Обработка ошибок десириализации
	if err != nil {
		responce := Response{Error: calculation.ErrInternalServer.Error()}
		encoder.Encode(responce)
		w.WriteHeader(http.StatusInternalServerError) // статус 500
		fmt.Fprintln(w, buf.String())
		log.Printf("[ERROR] Failed to decode json: %v", err)
		return
	}

	result, err := calculation.Calc(request.Expression)

	if err != nil {
		responce := Response{Error: err.Error()}
		encoder.Encode(responce)
		w.WriteHeader(http.StatusUnprocessableEntity) // статус 422 при невалидном JSON
		fmt.Fprintln(w, buf.String())
	} else {
		response := Response{Result: result}
		encoder.Encode(response)
		w.WriteHeader(http.StatusOK) // статус 200 при успешном вычислении
		fmt.Fprintln(w, buf.String())
	}
}

func (app *Application) RunServer() error {
	http.HandleFunc("/api/v1/calculate", CalcHandler)
	log.Printf("Server started at port %s", app.config.Addr)

	return http.ListenAndServe(":"+app.config.Addr, nil)
}
