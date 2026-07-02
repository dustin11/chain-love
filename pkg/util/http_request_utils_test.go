package util

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLimitRequestBodyBytesRejectsOversizedContentLength(t *testing.T) {
	req := httptest.NewRequest("POST", "/upload", strings.NewReader("123456"))
	req.ContentLength = 6
	recorder := httptest.NewRecorder()

	err := LimitRequestBodyBytes(recorder, req, 5)
	if !errors.Is(err, ErrRequestBodyTooLarge) {
		t.Fatalf("expected ErrRequestBodyTooLarge, got %v", err)
	}
}

func TestIsRequestBodyTooLargeError(t *testing.T) {
	if !IsRequestBodyTooLargeError(ErrRequestBodyTooLarge) {
		t.Fatal("expected sentinel error to be recognized")
	}
	if !IsRequestBodyTooLargeError(errors.New("http: request body too large")) {
		t.Fatal("expected max bytes reader error text to be recognized")
	}
}
