package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSessionExerciseAndAnswerFlow(t *testing.T) {
	handler := NewHandler()
	start := httptest.NewRecorder()
	handler.ServeHTTP(start, httptest.NewRequest(http.MethodPost, "/api/sessions", nil))
	if start.Code != http.StatusCreated {
		t.Fatalf("start status = %d, body = %s", start.Code, start.Body.String())
	}
	var session struct {
		ID string `json:"id"`
	}
	json.NewDecoder(start.Body).Decode(&session)

	next := httptest.NewRecorder()
	handler.ServeHTTP(next, httptest.NewRequest(http.MethodGet, "/api/sessions/"+session.ID+"/exercise", nil))
	if strings.Contains(next.Body.String(), "correctAnswer") || strings.Contains(next.Body.String(), "explanation") {
		t.Fatalf("exercise leaked grading data: %s", next.Body.String())
	}
	var exercise struct {
		ID string `json:"id"`
	}
	json.NewDecoder(next.Body).Decode(&exercise)

	answer := httptest.NewRecorder()
	body := strings.NewReader(`{"exerciseId":"` + exercise.ID + `","answer":"She walks to work every day."}`)
	handler.ServeHTTP(answer, httptest.NewRequest(http.MethodPost, "/api/sessions/"+session.ID+"/answers", body))
	if answer.Code != http.StatusOK || !strings.Contains(answer.Body.String(), `"correct":true`) {
		t.Fatalf("answer status = %d, body = %s", answer.Code, answer.Body.String())
	}
}

func TestCheckAnswerRejectsTrailingJSONWithoutChangingProgress(t *testing.T) {
	handler := NewHandler()
	start := httptest.NewRecorder()
	handler.ServeHTTP(start, httptest.NewRequest(http.MethodPost, "/api/sessions", nil))
	var session struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(start.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}

	next := httptest.NewRecorder()
	handler.ServeHTTP(next, httptest.NewRequest(http.MethodGet, "/api/sessions/"+session.ID+"/exercise", nil))
	var exercise struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(next.Body).Decode(&exercise); err != nil {
		t.Fatal(err)
	}

	answer := httptest.NewRecorder()
	body := strings.NewReader(`{"exerciseId":"` + exercise.ID + `","answer":"wrong"} {}`)
	handler.ServeHTTP(answer, httptest.NewRequest(http.MethodPost, "/api/sessions/"+session.ID+"/answers", body))
	if answer.Code != http.StatusBadRequest {
		t.Fatalf("answer status = %d, want %d; body = %s", answer.Code, http.StatusBadRequest, answer.Body.String())
	}

	progressResponse := httptest.NewRecorder()
	handler.ServeHTTP(progressResponse, httptest.NewRequest(http.MethodGet, "/api/sessions/"+session.ID+"/progress", nil))
	var progress struct {
		Hearts   int `json:"hearts"`
		XP       int `json:"xp"`
		Answered int `json:"answered"`
		Correct  int `json:"correct"`
	}
	if err := json.NewDecoder(progressResponse.Body).Decode(&progress); err != nil {
		t.Fatal(err)
	}
	if progress.Answered != 0 || progress.Correct != 0 || progress.XP != 0 || progress.Hearts != 5 {
		t.Fatalf("trailing JSON changed progress: %+v", progress)
	}
}

func TestCORSAllowsLocalDevelopmentFrontend(t *testing.T) {
	for _, origin := range []string{"http://localhost:3000", "http://localhost:3001"} {
		t.Run(origin, func(t *testing.T) {
			handler := NewHandler()
			request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
			request.Header.Set("Origin", origin)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if got := response.Header().Get("Access-Control-Allow-Origin"); got != origin {
				t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, origin)
			}
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
			}
		})
	}
}

func TestCORSRejectsUntrustedPreflight(t *testing.T) {
	handler := NewHandler()
	request := httptest.NewRequest(http.MethodOptions, "/api/sessions", nil)
	request.Header.Set("Origin", "https://attacker.example")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("preflight status = %d, want %d", response.Code, http.StatusForbidden)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("untrusted origin was allowed as %q", got)
	}
}

func TestCORSRejectsUntrustedStateChangingRequest(t *testing.T) {
	handler := NewHandler()
	request := httptest.NewRequest(http.MethodPost, "/api/sessions", nil)
	request.Header.Set("Origin", "https://attacker.example")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("cross-origin POST status = %d, want %d", response.Code, http.StatusForbidden)
	}
}
