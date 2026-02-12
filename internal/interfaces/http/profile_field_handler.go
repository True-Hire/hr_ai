package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type ProfileFieldHandler struct {
	service *application.ProfileFieldService
}

func NewProfileFieldHandler(service *application.ProfileFieldService) *ProfileFieldHandler {
	return &ProfileFieldHandler{service: service}
}

func (h *ProfileFieldHandler) Create(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
		return
	}

	var req CreateProfileFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	field, err := h.service.CreateProfileField(c.Request.Context(), userID, req.FieldName, req.SourceLang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create profile field"})
		return
	}

	c.JSON(http.StatusCreated, toProfileFieldResponse(field))
}

func (h *ProfileFieldHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid profile field id"})
		return
	}

	field, err := h.service.GetProfileField(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrProfileFieldNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "profile field not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get profile field"})
		return
	}

	c.JSON(http.StatusOK, toProfileFieldResponse(field))
}

func (h *ProfileFieldHandler) ListByUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
		return
	}

	fields, err := h.service.ListProfileFieldsByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list profile fields"})
		return
	}

	resp := make([]ProfileFieldResponse, 0, len(fields))
	for _, f := range fields {
		resp = append(resp, toProfileFieldResponse(&f))
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ProfileFieldHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid profile field id"})
		return
	}

	var req UpdateProfileFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	field, err := h.service.UpdateProfileField(c.Request.Context(), id, req.FieldName, req.SourceLang)
	if err != nil {
		if errors.Is(err, domain.ErrProfileFieldNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "profile field not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update profile field"})
		return
	}

	c.JSON(http.StatusOK, toProfileFieldResponse(field))
}

func (h *ProfileFieldHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid profile field id"})
		return
	}

	if err := h.service.DeleteProfileField(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrProfileFieldNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "profile field not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete profile field"})
		return
	}

	c.Status(http.StatusNoContent)
}
