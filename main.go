package main

import (
	"github.com/ddmendes/thumbnizer-demo/controller"
	"github.com/gin-gonic/gin"
)

func main() {
	control := controller.NewController()
	r := gin.Default()
	r.POST("/thumbnizer/", control.GetThumbnizerHandler())
	r.Run()
}
