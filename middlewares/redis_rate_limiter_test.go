package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	redismock "github.com/go-redis/redismock/v9"
)

func TestRedisRateLimiter_UnderLimit(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	mock.ExpectTxPipeline()
	mock.ExpectIncr("rate_limit:1.2.3.4:5678").SetVal(1)
	mock.ExpectExpire("rate_limit:1.2.3.4:5678", 10*time.Second).SetVal(true)
	mock.ExpectTxPipelineExec()
	mock.ExpectTTL("rate_limit:1.2.3.4:5678").SetVal(10 * time.Second)

	mw := RedisRateLimiter(db, 5, 10*time.Second)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "1.2.3.4:5678"
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)

	if rw.Code != 200 {
		t.Errorf("expected status 200, got %d", rw.Code)
	}
	if got := rw.Header().Get("X-RateLimit-Limit"); got != "5" {
		t.Errorf("expected X-RateLimit-Limit 5, got %q", got)
	}
	if got := rw.Header().Get("X-RateLimit-Remaining"); got != strconv.Itoa(4) {
		t.Errorf("expected X-RateLimit-Remaining 4, got %q", got)
	}
	if got := rw.Header().Get("X-RateLimit-Reset"); got == "" {
		t.Error("expected X-RateLimit-Reset to be set")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet redis expectations: %v", err)
	}
}

func TestRedisRateLimiter_OverLimit(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	mock.ExpectTxPipeline()
	mock.ExpectIncr("rate_limit:1.2.3.4:5678").SetVal(6)
	mock.ExpectExpire("rate_limit:1.2.3.4:5678", 10*time.Second).SetVal(true)
	mock.ExpectTxPipelineExec()
	mock.ExpectTTL("rate_limit:1.2.3.4:5678").SetVal(10 * time.Second)

	mw := RedisRateLimiter(db, 5, 10*time.Second)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called when over limit")
	}))
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "1.2.3.4:5678"
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)

	if rw.Code != 429 {
		t.Errorf("expected status 429, got %d", rw.Code)
	}
	if got := rw.Body.String(); got == "" || got == "\n" {
		t.Errorf("expected rate limit exceeded message, got %q", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet redis expectations: %v", err)
	}
}

func TestRedisRateLimiter_ExecError(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	mock.ExpectTxPipeline()
	mock.ExpectIncr("rate_limit:1.2.3.4:5678").SetVal(1)
	mock.ExpectExpire("rate_limit:1.2.3.4:5678", 10*time.Second).SetVal(true)
	mock.ExpectTxPipelineExec().SetErr(http.ErrAbortHandler)

	mw := RedisRateLimiter(db, 5, 10*time.Second)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called on exec error")
	}))
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "1.2.3.4:5678"
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)

	if rw.Code != 500 {
		t.Errorf("expected status 500, got %d", rw.Code)
	}
	if got := rw.Body.String(); got == "" || got == "\n" {
		t.Errorf("expected rate limit error message, got %q", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet redis expectations: %v", err)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		headers map[string]string
		remote  string
		want    string
		name    string
	}{
		{map[string]string{"X-Forwarded-For": "5.6.7.8"}, "1.2.3.4:5678", "5.6.7.8", "forwarded for"},
		{map[string]string{"X-Real-IP": "9.10.11.12"}, "1.2.3.4:5678", "9.10.11.12", "real ip"},
		{map[string]string{}, "8.8.8.8:1234", "8.8.8.8:1234", "remote addr"},
	}
	for _, tt := range tests {
		r := httptest.NewRequest("GET", "/", nil)
		for k, v := range tt.headers {
			r.Header.Set(k, v)
		}
		if tt.remote != "" {
			r.RemoteAddr = tt.remote
		}
		got := getClientIP(r)
		if got != tt.want {
			t.Errorf("%s: getClientIP() = %q, want %q", tt.name, got, tt.want)
		}
	}
}
