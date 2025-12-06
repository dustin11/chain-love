package test

import (
	"bytes"
	"chain-love/routers"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserSave(t *testing.T) {
	userName := "lisi"
	router := routers.SetupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/user/"+userName, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "用户"+userName+"已经保存", w.Body.String())
}

func TestUserRegisterForm(t *testing.T) {
	value := url.Values{}
	value.Add("email", "abcc@gmail.com")
	value.Add("password", "123456")
	value.Add("rePassword", "1234561")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/user/register", bytes.NewBufferString(value.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; param=value")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
