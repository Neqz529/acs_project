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

func DetectImageLabelsHandler(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get image file"})
		return
	}

	imageFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image file"})
		return
	}
	defer imageFile.Close()

	imageBytes, err := ioutil.ReadAll(imageFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image file"})
		return
	}

	tmpImageFile, err := ioutil.TempFile("", "image_*.jpeg")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp image file"})
		return
	}
	defer os.Remove(tmpImageFile.Name())
	tmpImageFile.Write(imageBytes)

	tmpVideoFile := fmt.Sprintf("%s.mp4", tmpImageFile.Name())

	cmd := exec.Command("ffmpeg", "-loop", "1", "-i", tmpImageFile.Name(), "-t", "2", "-vf", "fps=1", "-c:v", "libx264", "-pix_fmt", "yuv420p", tmpVideoFile)
	err = cmd.Run()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert image to video"})
		return
	}

	videoBytes, err := ioutil.ReadFile(tmpVideoFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read video file"})
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
		Features:     []videopb.Feature{videopb.Feature_LABEL_DETECTION},
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
		for _, label := range annotation.SegmentLabelAnnotations {
			for _, segment := range label.Segments {
				startTime := segment.Segment.StartTimeOffset.AsDuration().Seconds()
				endTime := segment.Segment.EndTimeOffset.AsDuration().Seconds()
				results = append(results, map[string]interface{}{
					"description": label.Entity.Description,
					"score":       segment.Confidence,
					"start":       startTime,
					"end":         endTime,
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"labels": results})
}
