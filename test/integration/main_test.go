package integration

import (
	"os"
	"testing"

	"github.com/daniel0321forever/terriyaki-go/test"
	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	test.Setup()
	exitCode := m.Run()
	test.Teardown()

	os.Exit(exitCode)
}
