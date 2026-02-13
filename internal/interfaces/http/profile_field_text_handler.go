package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type ProfileFieldTextHandler struct {
	service *application.ProfileFieldTextService
}

func NewProfileFieldTextHandler(service *application.ProfileFieldTextService) *ProfileFieldTextHandler {
	return &ProfileFieldTextHandler{service: service}
}

// Create godoc
// @Summary Create a text translation for a profile field
// @Tags profile-field-texts
// @Accept json
// @Produce json
// @Param id path string true "Profile field ID (UUID)"
// @Param request body CreateProfileFieldTextRequest true "Text data"
// @Success 201 {object} ProfileFieldTextResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /profile-fields/{id}/texts [post]
func (h *ProfileFieldTextHandler) Create(c *gin.Context) {
	fieldID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid profile field id"})
		return
	}

	var req CreateProfileFieldTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	text, err := h.service.CreateProfileFieldText(c.Request.Context(), req.ToDomain(fieldID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create profile field text"})
		return
	}

	c.JSON(http.StatusCreated, toProfileFieldTextResponse(text))
}

// Get godoc
// @Summary Get a profile field text by field ID and language
// @Tags profile-field-texts
// @Produce json
// @Param id path string true "Profile field ID (UUID)"
// @Param lang path string true "Language code (uz, ru, en)"
// @Success 200 {object} ProfileFieldTextResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /profile-fields/{id}/texts/{lang} [get]
func (h *ProfileFieldTextHandler) Get(c *gin.Context) {
	fieldID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid profile field id"})
		return
	}

	lang := c.Param("lang")
	if lang == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "lang is required"})
		return
	}

	text, err := h.service.GetProfileFieldText(c.Request.Context(), fieldID, lang)
	if err != nil {
		if errors.Is(err, domain.ErrProfileFieldTextNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "profile field text not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get profile field text"})
		return
	}

	c.JSON(http.StatusOK, toProfileFieldTextResponse(text))
}

// ListByField godoc
// @Summary List all text translations for a profile field
// @Tags profile-field-texts
// @Produce json
// @Param id path string true "Profile field ID (UUID)"
// @Success 200 {array} ProfileFieldTextResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /profile-fields/{id}/texts [get]
func (h *ProfileFieldTextHandler) ListByField(c *gin.Context) {
	fieldID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid profile field id"})
		return
	}

	texts, err := h.service.ListProfileFieldTexts(c.Request.Context(), fieldID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list profile field texts"})
		return
	}

	resp := make([]ProfileFieldTextResponse, 0, len(texts))
	for _, t := range texts {
		resp = append(resp, toProfileFieldTextResponse(&t))
	}

	c.JSON(http.StatusOK, resp)
}

// Update godoc
// @Summary Update a profile field text
// @Tags profile-field-texts
// @Accept json
// @Produce json
// @Param id path string true "Profile field ID (UUID)"
// @Param lang path string true "Language code (uz, ru, en)"
// @Param request body UpdateProfileFieldTextRequest true "Updated text data"
// @Success 200 {object} ProfileFieldTextResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /profile-fields/{id}/texts/{lang} [put]
func (h *ProfileFieldTextHandler) Update(c *gin.Context) {
	fieldID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid profile field id"})
		return
	}

	lang := c.Param("lang")
	if lang == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "lang is required"})
		return
	}

	var req UpdateProfileFieldTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	text, err := h.service.UpdateProfileFieldText(c.Request.Context(), req.ToDomain(fieldID, lang))
	if err != nil {
		if errors.Is(err, domain.ErrProfileFieldTextNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "profile field text not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update profile field text"})
		return
	}

	c.JSON(http.StatusOK, toProfileFieldTextResponse(text))
}

// Delete godoc
// @Summary Delete a profile field text
// @Tags profile-field-texts
// @Param id path string true "Profile field ID (UUID)"
// @Param lang path string true "Language code (uz, ru, en)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /profile-fields/{id}/texts/{lang} [delete]
func (h *ProfileFieldTextHandler) Delete(c *gin.Context) {
	fieldID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid profile field id"})
		return
	}

	lang := c.Param("lang")
	if lang == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "lang is required"})
		return
	}

	if err := h.service.DeleteProfileFieldText(c.Request.Context(), fieldID, lang); err != nil {
		if errors.Is(err, domain.ErrProfileFieldTextNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "profile field text not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete profile field text"})
		return
	}

	c.Status(http.StatusNoContent)
}
