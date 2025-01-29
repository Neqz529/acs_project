package endpoints

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/gin-gonic/gin"
	option "google.golang.org/api/option"
	visionpb "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

func DetectVideoHandler(c *gin.Context) {
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

	tmpFile, err := ioutil.TempFile("", "video-*.mp4")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temporary file"})
		return
	}
	defer os.Remove(tmpFile.Name())

	_, err = videoFile.Seek(0, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset video file pointer"})
		return
	}

	_, err = tmpFile.ReadFrom(videoFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save video to temporary file"})
		return
	}

	frameDir := "./frames"
	err = os.MkdirAll(frameDir, os.ModePerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory for frames"})
		return
	}

	cmd := exec.Command("ffmpeg", "-i", tmpFile.Name(), "-vf", "fps=1", fmt.Sprintf("%s/frame_%%03d.jpg", frameDir))
	err = cmd.Run()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract frames from video"})
		return
	}

	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx, option.WithCredentialsFile("/Users/andrej/Desktop/work_dir/go_projects/src/acs_project/sevice_key.json"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Vision API client"})
		return
	}
	defer client.Close()

	var results []map[string]interface{}

	err = filepath.Walk(frameDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".jpg" {
			frameFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer frameFile.Close()

			bytes, err := ioutil.ReadAll(frameFile)
			if err != nil {
				return err
			}

			image := &visionpb.Image{Content: bytes}
			labels, err := client.DetectLabels(ctx, image, nil, 10)
			if err != nil {
				return err
			}

			for _, label := range labels {
				results = append(results, map[string]interface{}{
					"description": label.Description,
					"score":       label.Score,
				})
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing frames"})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No labels detected"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"labels": results})
}
