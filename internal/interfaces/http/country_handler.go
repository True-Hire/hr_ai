package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CountryHandler struct {
	service *application.CountryService
}

func NewCountryHandler(service *application.CountryService) *CountryHandler {
	return &CountryHandler{service: service}
}

// GetByID godoc
// @Summary Get country by ID with translated name
// @Tags countries
// @Produce json
// @Param id path string true "Country ID (UUID)"
// @Param lang query string false "Language (uz, ru, en)" default(en)
// @Success 200 {object} CountryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /countries/{id} [get]
func (h *CountryHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid country id"})
		return
	}

	result, err := h.service.GetCountry(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrCountryNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "country not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get country"})
		return
	}

	lang := c.DefaultQuery("lang", "en")
	c.JSON(http.StatusOK, toCountryResponse(result, lang))
}

// List godoc
// @Summary List all countries with translated names
// @Tags countries
// @Produce json
// @Param lang query string false "Language (uz, ru, en)" default(en)
// @Success 200 {array} CountryResponse
// @Failure 500 {object} ErrorResponse
// @Router /countries [get]
func (h *CountryHandler) List(c *gin.Context) {
	results, err := h.service.ListCountries(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list countries"})
		return
	}

	lang := c.DefaultQuery("lang", "en")
	resp := make([]CountryResponse, 0, len(results))
	for _, cwt := range results {
		resp = append(resp, toCountryResponse(&cwt, lang))
	}

	c.JSON(http.StatusOK, resp)
}
