package endpoints

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	videointelligence "cloud.google.com/go/videointelligence/apiv1"
	"github.com/gin-gonic/gin"
	videopb "google.golang.org/genproto/googleapis/cloud/videointelligence/v1"
)

func DetectLogoFromImageHandler(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file"})
		return
	}

	imageFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer imageFile.Close()

	tmpImage, err := ioutil.TempFile("", "image-*.jpg")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp image file"})
		return
	}
	defer os.Remove(tmpImage.Name())

	_, err = ioutil.ReadAll(imageFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image"})
		return
	}

	err = c.SaveUploadedFile(file, tmpImage.Name())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	tmpVideo := fmt.Sprintf("%s.mp4", tmpImage.Name())
	defer os.Remove(tmpVideo)

	cmd := exec.Command("ffmpeg", "-loop", "1", "-i", tmpImage.Name(), "-t", "2", "-vf", "fps=1", "-c:v", "libx264", "-pix_fmt", "yuv420p", tmpVideo)
	err = cmd.Run()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert image to video"})
		return
	}

	videoBytes, err := ioutil.ReadFile(tmpVideo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read generated video"})
		return
	}

	ctx := context.Background()
	client, err := videointelligence.NewClient(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Video Intelligence API client"})
		return
	}
	defer client.Close()

	req := &videopb.AnnotateVideoRequest{
		InputContent: videoBytes,
		Features:     []videopb.Feature{videopb.Feature_LOGO_RECOGNITION},
	}

	operation, err := client.AnnotateVideo(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start video annotation"})
		return
	}

	resp, err := operation.Wait(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get annotation results"})
		return
	}

	var results []map[string]interface{}
	for _, annotation := range resp.AnnotationResults {
		for _, logo := range annotation.LogoRecognitionAnnotations {
			for _, track := range logo.Tracks { // Используем Tracks вместо Segments
				startTime := track.Segment.StartTimeOffset.AsDuration().Seconds()
				endTime := track.Segment.EndTimeOffset.AsDuration().Seconds()
				results = append(results, map[string]interface{}{
					"description": logo.Entity.Description,
					"score":       track.Confidence, // Используем Confidence из Track
					"start":       startTime,
					"end":         endTime,
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"logos": results})
}
