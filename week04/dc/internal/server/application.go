package server

import (
	"dc/api/common"
	appService "dc/internal/service"

	"github.com/gin-gonic/gin"
)

// CreateApplication 创建应用
func CreateApplication(c *gin.Context) {
	service := appService.DcApplicationService{}
	if err := c.ShouldBind(&service); err == nil {
		res := service.Post(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, common.ErrorResponse(err))
	}
}
