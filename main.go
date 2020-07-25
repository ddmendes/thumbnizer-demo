package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

type ThumbnizerPayload struct {
	UUID    string
	Picture string
	Output  string
}

func main() {
	r := gin.Default()
	r.POST("/thumbnizer/", thumbnizerHandler)
	r.Run()
}

func thumbnizerHandler(c *gin.Context) {
	var payload ThumbnizerPayload
	err := c.BindJSON(&payload)
	if err != nil {
		log.Panic(err)
	}
	log.Println(payload)
	c.JSON(200, payload)
}
