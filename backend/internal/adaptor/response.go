package adaptor

import "github.com/gin-gonic/gin"

func Success(c *gin.Context, status int, payload any) {
	c.JSON(status, payload)
}

func Error(c *gin.Context, status int, code string, extras gin.H) {
	errBody := gin.H{"code": code}
	for k, v := range extras {
		errBody[k] = v
	}
	c.JSON(status, gin.H{"error": errBody})
}
