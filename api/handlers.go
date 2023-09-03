package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func HealthCheck(c *gin.Context) {
	r := map[string]string{"online": "true"}
	c.JSON(http.StatusOK, r)
}
