package main

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	// 实例化一个Echo类型
	e := echo.New()
	// 添加中间件
	e.Use(middleware.Logger())
	// 添加路由
	e.GET("/", hello)
	// 启动服务器
	e.Logger.Fatal(e.Start(":1323"))
}

func hello(c echo.Context) error {
	return c.JSON(http.StatusOK, "Hello, World !")
}
