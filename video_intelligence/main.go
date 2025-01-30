package main

import (
	"os"

	"github.com/gin-gonic/gin"

	endpoints "video_intelligence/endpoints"
)

func main() {
	r := gin.Default()
	r.POST("/detect_video", endpoints.DetectVideoLabelsHandler)
	r.POST("/detect_logo", endpoints.DetectLogoFromImageHandler)
	r.POST("/detect_object", endpoints.DetectImageLabelsHandler)
	port := ":" + os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}
	r.Run(port)
}
