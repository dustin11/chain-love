package test

import (
	"chain-love/domain/ds/enum"
	r "chain-love/routers"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var router *gin.Engine

func init() {
	gin.SetMode(gin.TestMode)
	router = r.SetupRouter()
}

func TestIndexGetRouter(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "hello gin get method", "返回的HTML页面中应该包含 hello gin get method")
}

func TestIndexPostRouter(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hello gin post method", w.Body.String())
}

func TestEnum(t *testing.T) {
	var dd = enum.Decoration_Simple
	sss := enum.GetList(dd)
	fmt.Println(sss)

	//reflect3.TypeByName("")
	//t := reflect2.TypeByName("enum.DecorationType")
	//t.Type1().Name()
}

func TestGetIp(t *testing.T) {
	//ip := exnet.ClientPublicIP(ctx.Request)
	//if ip == ""{
	//	ip = exnet.ClientIP(ctx.Request)
	//}
}

func TestIdWorker(t *testing.T) {

	//snowflake.Epoch = 1604304591346//time.Now().UnixNano() / int64(1e6)
	snowflake.NodeBits = 2
	snowflake.StepBits = 5
	n, err := snowflake.NewNode(1)
	if err != nil {
		println(err)
		os.Exit(1)
	}

	for i := 0; i < 3; i++ {
		id := n.Generate()
		fmt.Println("id", id)
		fmt.Println(
			"node: ", id.Node(),
			"step: ", id.Step(),
			"time: ", id.Time(),
			"\n\t",
		)
	}
}
