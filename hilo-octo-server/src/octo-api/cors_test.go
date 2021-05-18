package main

import (
	"net/http"
	"testing"
	"time"

	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

var cors = &CORS{
	AllowAllOrigins: true,
	AllowHeaders:    []string{"X-Octo-Key", "accept"},
	AllowMethods:    []string{"GET", "POST"},
	MaxAge:          1 * time.Minute,
}

func TestPanic(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router := gin.New()
	c := &CORS{
		AllowAllOrigins: false,
		AllowMethods:    []string{"GET"},
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	router.Use(c.MiddleWare())

	router.ServeHTTP(w, req)

}

func TestNoOrigin(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router := gin.New()
	c := &CORS{
		AllowAllOrigins: false,
		AllowMethods:    []string{"GET"},
		AllowOrigins:    []string{"http://octotest.com"},
	}
	router.Use(c.MiddleWare())

	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("This should not Match")
	}
}

func TestNotAllowOrigin(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router := gin.New()
	c := &CORS{
		AllowAllOrigins: false,
		AllowMethods:    []string{"GET"},
		AllowOrigins:    []string{"http://octotest.com"},
	}

	req.Header.Set("Origin", "http://localhost.com")

	router.Use(c.MiddleWare())

	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("This should not Match")
	}
}

func TestPreFlightRequest(t *testing.T) {
	req, _ := http.NewRequest("OPTIONS", "/", nil)
	w := httptest.NewRecorder()

	req.Header.Set("Origin", "http://octotest.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "X-Octo-Key")
	router := gin.New()

	router.Use(cors.MiddleWare())
	router.ServeHTTP(w, req)
	if w.Header().Get("Access-Control-Allow-Methods") != "GET" {
		t.Fatal("Mismatch of methods.")
	}
}

func TestPreFlightRequestWithNotMethod(t *testing.T) {
	req, _ := http.NewRequest("OPTIONS", "/", nil)
	w := httptest.NewRecorder()

	req.Header.Set("Origin", "http://octotest.com")
	req.Header.Set("Access-Control-Request-Method", "PUT")
	req.Header.Set("Access-Control-Request-Headers", "X-Octo-Key")
	router := gin.New()

	router.Use(cors.MiddleWare())
	router.ServeHTTP(w, req)
	if w.Header().Get("Access-Control-Allow-Methods") != "" {
		t.Fatal("Match of methods.")
	}
}

func TestInValidHeader(t *testing.T) {
	req, _ := http.NewRequest("OPTIONS", "/", nil)
	w := httptest.NewRecorder()

	req.Header.Set("Origin", "http://localhost")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "X-Octo-Key")

	router := gin.New()

	c := &CORS{
		AllowAllOrigins:  false,
		AllowCredentials: true,
		AllowOrigins:     []string{"http://localhost"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "Accept"},
		AllowMethods:     []string{"GET", "POST"},
		MaxAge:           1 * time.Minute,
	}

	router.Use(c.MiddleWare())

	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Headers") != "" {
		t.Fatal("Vaild Header!")
	}
}

func TestNoValidHeader(t *testing.T) {
	req, _ := http.NewRequest("OPTIONS", "/", nil)
	w := httptest.NewRecorder()

	req.Header.Set("Origin", "http://localhost")
	req.Header.Set("Access-Control-Request-Method", "GET")

	router := gin.New()

	c := &CORS{
		AllowAllOrigins:  false,
		AllowCredentials: true,
		AllowOrigins:     []string{"http://localhost"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "Accept"},
		AllowMethods:     []string{"GET", "POST"},
		MaxAge:           1 * time.Minute,
	}

	router.Use(c.MiddleWare())

	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Headers") != "" {
		t.Fatal("Vaild Header!")
	}
}

func TestCredential(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	req.Header.Set("Origin", "http://localhost")

	router := gin.New()

	c := &CORS{
		AllowAllOrigins:  false,
		AllowCredentials: true,
		AllowOrigins:     []string{"http://localhost"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Authorization"},
		AllowMethods:     []string{"GET", "POST"},
		MaxAge:           1 * time.Minute,
	}

	router.Use(c.MiddleWare())

	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost" {
		t.Fatal("Invaild Origin is set")
	}

	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatal("Invaild Credentials is set")
	}
}
func TestPreFlightNotAllAllowOrigin(t *testing.T) {
	req, _ := http.NewRequest("OPTIONS", "/", nil)
	w := httptest.NewRecorder()

	req.Header.Set("Origin", "http://localhost")
	req.Header.Set("Access-Control-Request-Headers", "X-Octo-Key")
	req.Header.Set("Access-Control-Request-Method", "GET")

	router := gin.New()

	c := &CORS{
		AllowAllOrigins: false,
		AllowMethods:    []string{"GET"},
		AllowOrigins:    []string{"http://localhost"},
		AllowHeaders:    []string{"X-Octo-Key", "accept"},
	}

	router.Use(c.MiddleWare())

	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost" {
		t.Fatal("Invalid AllowOrigins")
	}
}

func TestAllAllowOrigin(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	req.Header.Set("Origin", "http://localhost")

	router := gin.New()

	c := &CORS{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET"},
	}

	router.Use(c.MiddleWare())

	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("Invalid AllowAllOrigins")
	}
}

func TestFalseMethodValidate(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", nil)

	w := httptest.NewRecorder()

	req.Header.Set("Origin", "http://localhost")

	router := gin.New()

	c := &CORS{
		AllowAllOrigins: false,
		AllowOrigins:    []string{"http://localhost"},
		AllowMethods:    []string{"GET"},
	}

	router.Use(c.MiddleWare())

	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Methods") != "" {
		t.Fatal("Invalid MethodValidate")
	}
}

func TestAllHeadersValidate(t *testing.T) {
	req, _ := http.NewRequest("OPTIONS", "/", nil)

	w := httptest.NewRecorder()

	req.Header.Set("Origin", "http://localhost")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "X-Octo-Key,Origin")

	router := gin.New()
	c := &CORS{
		AllowAllOrigins: false,
		AllowMethods:    []string{"GET", "POST"},
		AllowOrigins:    []string{"http://localhost"},
		AllowHeaders:    []string{"X-Octo-Key", "Origin", "accept"},
	}
	router.Use(c.MiddleWare())

	router.ServeHTTP(w, req)
	t.Log(w.Header().Get("Access-Control-Allow-Headers"))
	if w.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Fatal("InVaild Header!")
	}
}
