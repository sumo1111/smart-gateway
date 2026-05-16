package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sumo1111/smart-gateway/common"
	"github.com/sumo1111/smart-gateway/model"
)

// TokenAuth API令牌认证
func TokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			// 也检查查询参数
			auth = "Bearer " + c.Query("key")
		}
		if !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			c.Abort()
			return
		}
		key := strings.TrimPrefix(auth, "Bearer ")
		token, err := model.GetTokenByKey(key)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Set("token", token)
		c.Next()
	}
}

// AdminAuth 管理员认证
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		password := c.GetHeader("X-Admin-Password")
		if password == "" {
			// 也检查Bearer token
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				password = strings.TrimPrefix(auth, "Bearer ")
			}
		}
		if password != common.AdminPassword {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "admin authentication required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
