package main

import (
	"main/page"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")
	router.GET("/", page.Game)
	router.POST("/", page.Result)
	router.Run()
}
