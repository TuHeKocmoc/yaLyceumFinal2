package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalcHandler(t *testing.T) {
	sendRequest := func(method, url string, body []byte) (*httptest.ResponseRecorder, error) {
		req := httptest.NewRequest(method, url, bytes.NewReader(body))
		w := httptest.NewRecorder()

		CalcHandler(w, req)
		return w, nil
	}

	t.Run("Valid JSON, valid expression", func(t *testing.T) {
		reqBody := Request{Expression: "2+2"}
		jsonBody, _ := json.Marshal(reqBody)

		w, _ := sendRequest("POST", "/api/v1/calculate", jsonBody)
		if w.Result().StatusCode != http.StatusOK {
			t.Fatalf("Status=%d; want %d", w.Result().StatusCode, http.StatusOK)
		}

		var out Output
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatalf("Unmarshal error: %v, body=%s", err, w.Body.String())
		}
		if out.Result != 4 {
			t.Errorf("Got result=%f; want %f", out.Result, 4.0)
		}
	})

	t.Run("Valid JSON, invalid expression (буква)", func(t *testing.T) {
		reqBody := Request{Expression: "2+2a"}
		jsonBody, _ := json.Marshal(reqBody)

		w, _ := sendRequest("POST", "/api/v1/calculate", jsonBody)
		if w.Result().StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("Status=%d; want %d", w.Result().StatusCode, http.StatusUnprocessableEntity)
		}

		var errResp Err
		if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if errResp.Error != "Expression is not valid" {
			t.Errorf("Got error=%q; want %q", errResp.Error, "Expression is not valid")
		}
	})

	t.Run("Valid JSON, leads to internal error (деление на ноль)", func(t *testing.T) {
		reqBody := Request{Expression: "10/0"}
		jsonBody, _ := json.Marshal(reqBody)

		w, _ := sendRequest("POST", "/api/v1/calculate", jsonBody)
		if w.Result().StatusCode != http.StatusInternalServerError {
			t.Fatalf("Status=%d; want %d", w.Result().StatusCode, http.StatusInternalServerError)
		}

		var errResp Err
		if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if errResp.Error != "Internal server error" {
			t.Errorf("Got error=%q; want %q", errResp.Error, "Internal server error")
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		invalidBody := []byte(`{ expression: 2+2 }`)
		w, _ := sendRequest("POST", "/api/v1/calculate", invalidBody)
		if w.Result().StatusCode != http.StatusInternalServerError {
			t.Fatalf("Status=%d; want %d", w.Result().StatusCode, http.StatusInternalServerError)
		}
		var errResp Err
		if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if errResp.Error != "Internal server error" {
			t.Errorf("Got error=%q; want %q", errResp.Error, "Internal server error")
		}
	})
}
