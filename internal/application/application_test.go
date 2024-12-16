package application

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalcHandlerBadCase(t *testing.T) {
	testCases := []struct {
		name           string
		expression     string
		expectedBody   string
		expectedStatus int
	}{
		{
			name:           "Empty JSON",
			expression:     ``,
			expectedBody:   `{"error":"internal server error"}`,
			expectedStatus: 500,
		},
		{
			name:           "No expression field",
			expression:     `{"sometext": "2+2"}`,
			expectedBody:   `{"error":"empty expression"}`,
			expectedStatus: 422,
		},
		{
			name:           "Empty expression",
			expression:     `{"expression": ""}`,
			expectedBody:   `{"error":"empty expression"}`,
			expectedStatus: 422,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBuffer([]byte(tc.expression)))
			w := httptest.NewRecorder()
			CalcHandler(w, req)
			res := w.Result()

			var responceBody map[string]string
			json.NewDecoder(res.Body).Decode(&responceBody)
			body, _ := json.Marshal(responceBody)

			if tc.expectedBody != string(body) {
				t.Errorf("expected body %s, got %s", tc.expectedBody, body)
			}

			if res.StatusCode != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, res.StatusCode)
			}
		})
	}
}

func TestCalcHandlerSuccessCase(t *testing.T) {
	testCases := []struct {
		name           string
		expression     string
		expectedBody   string
		expectedStatus int
	}{
		{
			name:           "Simple expression",
			expression:     `{"expression": "2+2*2"}`,
			expectedBody:   `{"result":6}`,
			expectedStatus: 200,
		},
		{
			name:           "Hard expression",
			expression:     `{"expression": "(52-49)*4-1"}`,
			expectedBody:   `{"result":11}`,
			expectedStatus: 200,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBuffer([]byte(tc.expression)))
			w := httptest.NewRecorder()
			CalcHandler(w, req)
			res := w.Result()

			var responceBody map[string]int
			json.NewDecoder(res.Body).Decode(&responceBody)
			body, _ := json.Marshal(responceBody)

			if tc.expectedBody != string(body) {
				t.Errorf("expected body %s, got %s", tc.expectedBody, body)
			}

			if res.StatusCode != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, res.StatusCode)
			}
		})
	}
}
