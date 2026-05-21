package integration

import (
	"os"
	"testing"

	"log"
	"net/http"
	"time"

	"github.com/daniel0321forever/terriyaki-go/tests"
	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	tests.Setup()
	// start backend server in-process for E2E tests
	router := tests.Router()
	srvErrCh := make(chan error, 1)
	go func() {
		if err := router.Run(":8080"); err != nil {
			srvErrCh <- err
		}
	}()

	// wait for server to become ready
	ready := false
	for i := 0; i < 20; i++ {
		resp, err := http.Get("http://localhost:8080/api/v1/ping")
		if err == nil && resp.StatusCode == 200 {
			ready = true
			break
		}
		select {
		case err := <-srvErrCh:
			log.Printf("server failed to start: %v", err)
			ready = false
			i = 999
		default:
			time.Sleep(200 * time.Millisecond)
		}
	}
	if !ready {
		log.Printf("warning: backend server did not become ready; tests may fail")
	}
	exitCode := m.Run()
	tests.Teardown()

	os.Exit(exitCode)
}
