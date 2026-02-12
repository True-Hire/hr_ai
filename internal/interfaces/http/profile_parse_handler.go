package http

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

var allowedMimeTypes = map[string]bool{
	"application/pdf": true,
	"image/png":       true,
	"image/jpeg":      true,
	"text/plain":      true,
}

const maxFileSize = 20 << 20 // 20 MB

type ProfileParseHandler struct {
	service *application.ProfileParseService
}

func NewProfileParseHandler(service *application.ProfileParseService) *ProfileParseHandler {
	return &ProfileParseHandler{service: service}
}

func (h *ProfileParseHandler) Parse(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
		return
	}

	contentType := c.ContentType()

	if strings.HasPrefix(contentType, "multipart/form-data") {
		h.handleMultipart(c, userID)
		return
	}

	h.handleJSON(c, userID)
}

func (h *ProfileParseHandler) handleJSON(c *gin.Context, userID uuid.UUID) {
	var req ProfileParseTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	result, err := h.service.ParseFromText(c.Request.Context(), userID, req.UserInput)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
			return
		}
		log.Printf("profile parse error: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to parse profile"})
		return
	}

	c.JSON(http.StatusOK, toProfileParseResponse(result))
}

func (h *ProfileParseHandler) handleMultipart(c *gin.Context, userID uuid.UUID) {
	userInput := c.PostForm("user_input")
	if userInput != "" {
		result, err := h.service.ParseFromText(c.Request.Context(), userID, userInput)
		if err != nil {
			if errors.Is(err, domain.ErrUserNotFound) {
				c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to parse profile"})
			return
		}
		c.JSON(http.StatusOK, toProfileParseResponse(result))
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "either user_input or file is required"})
		return
	}

	if fileHeader.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file too large (max 20MB)"})
		return
	}

	mimeType := fileHeader.Header.Get("Content-Type")
	if !allowedMimeTypes[mimeType] {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "unsupported file type; allowed: PDF, PNG, JPG/JPEG, TXT",
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read uploaded file"})
		return
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read file data"})
		return
	}

	result, err := h.service.ParseFromFile(c.Request.Context(), userID, fileData, mimeType)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
			return
		}
		log.Printf("profile parse from file error: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to parse profile from file"})
		return
	}

	c.JSON(http.StatusOK, toProfileParseResponse(result))
}
