package appRouter

import (
	appServer "dc/internal/server"

	"github.com/gin-gonic/gin"
)

func RegisterAppRouter(r *gin.Engine) {
	datacenterApi := r.Group("/api/datacenter/v1")
	{
		// 应用管理
		datacenterApi.POST("/applications", appServer.CreateApplication) // 应用创建
		datacenterApi.POST("alive", Alive)
	}
}

// Alive 保活
func Alive(c *gin.Context) {
	c.String(200, "ok")
}
