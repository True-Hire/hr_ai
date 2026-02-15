package http

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ruziba3vich/hr-ai/internal/application"
)

type StorageHandler struct {
	service *application.StorageService
}

func NewStorageHandler(service *application.StorageService) *StorageHandler {
	return &StorageHandler{service: service}
}

// Upload godoc
// @Summary Upload a file to storage
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Param folder formData string true "Storage folder prefix (e.g. profile-photos, resumes)"
// @Param file formData file true "File to upload"
// @Success 201 {object} UploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /files [post]
func (h *StorageHandler) Upload(c *gin.Context) {
	folder := c.PostForm("folder")
	if folder == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "folder is required"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to open file"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read file"})
		return
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	result, err := h.service.Upload(c.Request.Context(), folder, data, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to upload file"})
		return
	}

	c.JSON(http.StatusCreated, UploadResponse{
		ObjectName: result.ObjectName,
		URL:        result.URL,
	})
}

// Get godoc
// @Summary Download a file from storage
// @Tags files
// @Produce octet-stream
// @Param object_name query string true "Object name (e.g. profile-photos/uuid)"
// @Success 200 {file} binary
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /files [get]
func (h *StorageHandler) Get(c *gin.Context) {
	objectName := c.Query("object_name")
	if objectName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "object_name is required"})
		return
	}

	info, err := h.service.Get(c.Request.Context(), objectName)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "file not found"})
		return
	}

	c.Header("Content-Type", info.ContentType)
	c.Header("Content-Length", strconv.FormatInt(info.Size, 10))
	c.Data(http.StatusOK, info.ContentType, info.Data)
}

// Delete godoc
// @Summary Delete a file from storage
// @Tags files
// @Produce json
// @Param object_name query string true "Object name (e.g. profile-photos/uuid)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /files [delete]
func (h *StorageHandler) Delete(c *gin.Context) {
	objectName := c.Query("object_name")
	if objectName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "object_name is required"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), objectName); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete file"})
		return
	}

	c.Status(http.StatusNoContent)
}

type UploadResponse struct {
	ObjectName string `json:"object_name"`
	URL        string `json:"url"`
}

