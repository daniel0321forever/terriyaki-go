package integration

import (
	"os"
	"testing"

	"github.com/daniel0321forever/terriyaki-go/tests"
	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	tests.Setup()
	exitCode := m.Run()
	tests.Teardown()

	os.Exit(exitCode)
}
