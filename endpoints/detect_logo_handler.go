package endpoints

import (
	"context"
	"io/ioutil"
	"net/http"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/gin-gonic/gin"
	option "google.golang.org/api/option"
	visionpb "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

func DetectLogoHandler(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file"})
		return
	}

	imageData, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer imageData.Close()

	bytes, err := ioutil.ReadAll(imageData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx, option.WithCredentialsFile("/Users/andrej/Desktop/work_dir/go_projects/src/acs_project/sevice_key.json"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Vision API client"})
		return
	}
	defer client.Close()

	image := &visionpb.Image{Content: bytes}
	logos, err := client.DetectLogos(ctx, image, nil, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to detect logos"})
		return
	}

	var results []map[string]interface{}
	for _, logo := range logos {
		results = append(results, map[string]interface{}{
			"description": logo.Description,
			"score":       logo.Score,
		})
	}

	c.JSON(http.StatusOK, gin.H{"logos": results})
}
