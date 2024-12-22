package main

import (
	"encoding/json"
	"net/http"
)

type Output struct {
	Result float64 `json:"result"`
}
type Err struct {
	Error string `json:"error"`
}

type Request struct {
	Expression string `json:"expression"`
}

func CalcHandler(w http.ResponseWriter, r *http.Request) {
	var request Request
	// получаем данные из запроса
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		output := Err{Error: "Internal server error"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(output)
		return
	}
	// проверяем валидность введенных данных
	check := checkInput(request.Expression)
	if !check {
		w.WriteHeader(http.StatusUnprocessableEntity)
		output := Err{Error: "Expression is not valid"}
		json.NewEncoder(w).Encode(output)
		return
	}
	// считаем результат
	result, err := Calc(request.Expression)
	if err != nil {
		output := Err{Error: "Internal server error"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(output)
		return
	}
	// отправляем результат
	output := Output{Result: result}
	json.NewEncoder(w).Encode(output)
}

func main() {
	// запускаем сервер
	http.HandleFunc("/api/v1/calculate", CalcHandler)
	http.ListenAndServe(":8080", nil)
}
