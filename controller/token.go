package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sumo1111/smart-gateway/model"
)

func GetAllTokens(c *gin.Context) {
	tokens, err := model.GetAllTokens()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tokens})
}

func AddToken(c *gin.Context) {
	var t model.Token
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if t.Status == 0 { t.Status = 1 }
	if err := model.CreateToken(&t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": t})
}

func UpdateToken(c *gin.Context) {
	var t model.Token
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := model.UpdateToken(&t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": t})
}

func DeleteToken(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := model.DeleteToken(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func RefreshTokenKey(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	// 生成新key
	t := &model.Token{ID: id, Key: "sk-" + generateKey()}
	// 直接更新key
	model.DB.Exec("UPDATE tokens SET key=? WHERE id=?", t.Key, id)
	c.JSON(http.StatusOK, gin.H{"key": t.Key})
}

func generateKey() string {
	return strconv.FormatInt(idCounter(), 36) + "sgw" + randomHex(12)
}

func idCounter() int64 {
	var count int64
	model.DB.QueryRow("SELECT COUNT(*) FROM tokens").Scan(&count)
	return count + 1
}

func randomHex(n int) string {
	// 简易随机字符串
	const chars = "0123456789abcdef"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[i%16]
	}
	return string(b)
}
