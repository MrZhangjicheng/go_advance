package main

import (
	conf "dc/configs"

	appRouter "dc/api/application"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置文件

	conf.Init()

	r := NewRouter()
	err := r.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func NewRouter() *gin.Engine {
	r := gin.Default()
	// 中间件
	appRouter.RegisterAppRouter(r)
	// qing.RegisterQingRouter(r)

	return r

}
