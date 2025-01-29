package endpoints

import (
	"context"
	"io/ioutil"
	"net/http"

	config "cloud_vision/config"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/gin-gonic/gin"
	option "google.golang.org/api/option"
	visionpb "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

func DetectObjectHandler(c *gin.Context) {
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
	client, err := vision.NewImageAnnotatorClient(ctx, option.WithCredentialsFile(config.GCPCredentials))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Vision API client"})
		return
	}
	defer client.Close()

	image := &visionpb.Image{Content: bytes}
	labels, err := client.DetectLabels(ctx, image, nil, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to detect objects"})
		return
	}

	var results []map[string]interface{}
	for _, label := range labels {

		results = append(results, map[string]interface{}{
			"name":  label.Description,
			"score": label.Score,
		})

	}

	if len(results) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Object not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"objects": results})
}
