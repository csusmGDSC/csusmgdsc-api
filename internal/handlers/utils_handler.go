package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/csusmGDSC/csusmgdsc-api/config"
	"github.com/labstack/echo/v4"
)

// uploadImage handles the image upload route.
//
// The request must contain a FormFile named "image". The image is uploaded to
// an S3 bucket and the URL of the uploaded image is returned as a JSON response.
//
// The response is a JSON object with a single key "url" which contains the URL
// of the uploaded image.
func (h *Handler) UploadImage(c echo.Context) error {
	userRole, ok := c.Get("user_role").(string)
	if !ok || userRole != "ADMIN" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Insufficient permissions"})
	}

	file, err := c.FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid file request. Send file as FormFile."})
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not read image."})
	}
	defer src.Close()

	cfg := config.LoadConfig()

	sess, err := session.NewSession(&aws.Config{
		Region:      &cfg.AWSRegion,
		Credentials: credentials.NewStaticCredentials(cfg.AWSAccessKey, cfg.AWSSecretAccessKey, ""),
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not connect to the image storage."})
	}

	s3Svc := s3.New(sess)
	fileName := fmt.Sprintf("%d%s", file.Size, filepath.Ext(file.Filename))

	_, err = s3Svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(cfg.S3BucketName),
		Key:         aws.String(fileName),
		Body:        src,
		ContentType: aws.String(file.Header.Get("Content-Type")),
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not upload image.", "message": err.Error()})
	}

	imageURL := fmt.Sprintf("%s%s", cfg.AWSCloudfrontDomain, fileName)
	return c.JSON(http.StatusOK, map[string]string{"url": imageURL})
}

// RemoveImage handles the image deletion route.
//
// The request must contain a query parameter "url" which contains the URL of the
// image to be deleted.
//
// The response is a JSON object with a single key "message" which contains the
// message of the deletion result.
func (h *Handler) RemoveImage(c echo.Context) error {
	userRole, ok := c.Get("user_role").(string)
	if !ok || userRole != "ADMIN" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Insufficient permissions"})
	}

	imageURL := c.QueryParam("url")
	if imageURL == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Image URL is required"})
	}

	cfg := config.LoadConfig()

	// Extract the file key from the image URL
	s3Prefix := cfg.AWSCloudfrontDomain
	if !strings.HasPrefix(imageURL, s3Prefix) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid image URL"})
	}
	imageKey := strings.TrimPrefix(imageURL, s3Prefix)

	sess, err := session.NewSession(&aws.Config{
		Region:      &cfg.AWSRegion,
		Credentials: credentials.NewStaticCredentials(cfg.AWSAccessKey, cfg.AWSSecretAccessKey, ""),
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not connect to the image storage."})
	}

	s3Svc := s3.New(sess)

	_, err = s3Svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(cfg.S3BucketName),
		Key:    aws.String(imageKey),
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not delete image"})
	}

	err = s3Svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(cfg.S3BucketName),
		Key:    aws.String(imageKey),
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Image deletion verification failed"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Image deleted successfully"})
}
