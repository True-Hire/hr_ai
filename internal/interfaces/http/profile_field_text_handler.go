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
