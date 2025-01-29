package main

import (
	"github.com/gin-gonic/gin"

	endpoints "cloud_vision/endpoints"
)

func main() {
	r := gin.Default()
	r.POST("/detect_logo", endpoints.DetectLogoHandler)
	r.POST("/detect_video", endpoints.DetectVideoHandler)
	r.POST("/detect_object", endpoints.DetectObjectHandler)
	r.Run(":8080")
}
