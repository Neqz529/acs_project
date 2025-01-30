package main

import (
	"os"

	"github.com/gin-gonic/gin"

	endpoints "cloud_vision/endpoints"
)

func main() {
	r := gin.Default()
	r.POST("/detect_logo", endpoints.DetectLogoHandler)
	r.POST("/detect_video", endpoints.DetectVideoHandler)
	r.POST("/detect_object", endpoints.DetectObjectHandler)

	port := ":" + os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":8080")
}
