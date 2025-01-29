package main

import (
	"github.com/gin-gonic/gin"

	endpoints "video_intelligence/endpoints"
)

func main() {
	r := gin.Default()
	r.POST("/detect_video", endpoints.DetectVideoLabelsHandler)
	r.POST("/detect_logo", endpoints.DetectLogoFromImageHandler)
	r.POST("/detect_object", endpoints.DetectImageLabelsHandler)
	r.Run(":8080")
}
