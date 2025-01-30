package endpoints

import (
	"context"
	"io/ioutil"
	"net/http"

	videointelligence "cloud.google.com/go/videointelligence/apiv1"
	"github.com/gin-gonic/gin"
	videopb "google.golang.org/genproto/googleapis/cloud/videointelligence/v1"
)

func DetectVideoLabelsHandler(c *gin.Context) {
	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get video file"})
		return
	}

	videoFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open video file"})
		return
	}
	defer videoFile.Close()

	videoBytes, err := ioutil.ReadAll(videoFile)
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
