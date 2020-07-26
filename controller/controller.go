package controller

import (
	"bufio"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"runtime"

	"github.com/ddmendes/thumbnizer-demo/pipeline"
	"github.com/gin-gonic/gin"
)

// Controller is an access point to thumbnizer handlers.
type Controller struct {
	pipeline *pipeline.Pipeline
}

type ThumbnizerRequest struct {
	UUID    string
	Picture string
	Output  string
}

type ThumbnizerResponse struct {
	UUID    string `json:uuid`
	Status  string `json:status`
	Message string `json:message,omitempty`
}

var extJPEG = "jpg"

// NewController creates a new controller.
func NewController() *Controller {
	p := pipeline.Boot(500, runtime.NumCPU())
	return &Controller{p}
}

func (c *Controller) GetThumbnizerHandler() func(*gin.Context) {
	return func(ctx *gin.Context) {
		var payload ThumbnizerRequest
		err := ctx.BindJSON(&payload)
		if err != nil {
			log.Println(err)
			ctx.JSON(400, err)
			return
		}
		job := &pipeline.Job{
			UUID:        payload.UUID,
			Read:        readFunc(payload.Picture),
			WriteSmall:  writeFunc(payload.Output, "small", extJPEG),
			WriteMedium: writeFunc(payload.Output, "medium", extJPEG),
			WriteLarge:  writeFunc(payload.Output, "large", extJPEG),
			Callback:    responseFunc(ctx),
		}

		c.pipeline.Push(job)
	}
}

func readFunc(path string) func() (image.Image, error) {
	return func() (image.Image, error) {
		f, err := os.Open(path)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer f.Close()

		bufReader := bufio.NewReader(f)
		image, _, err := image.Decode(bufReader)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		return image, nil
	}
}

func writeFunc(path, suffix, ext string) func(image.Image) error {
	return func(image image.Image) error {
		fpath := fmt.Sprintf("%s_%s.%s", path, suffix, ext)

		f, err := os.Open(fpath)
		if err != nil {
			log.Println(err)
			return err
		}
		defer f.Close()

		bufWriter := bufio.NewWriter(f)
		err = jpeg.Encode(bufWriter, image, nil)
		if err != nil {
			log.Println(err)
		}
		return err
	}
}

func responseFunc(ctx *gin.Context) func(*pipeline.Job) {
	return func(job *pipeline.Job) {
		var response ThumbnizerResponse
		var code int
		response.UUID = job.UUID
		if job.Err != nil {
			code = 500
			response.Status = "failed"
			response.Message = job.Err.Error()
		} else {
			code = 200
			response.Status = "success"
		}
		ctx.JSON(code, response)
	}
}
