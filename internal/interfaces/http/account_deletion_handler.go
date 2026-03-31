package http

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type AccountDeletionHandler struct {
	service *application.AccountDeletionService
}

func NewAccountDeletionHandler(service *application.AccountDeletionService) *AccountDeletionHandler {
	return &AccountDeletionHandler{service: service}
}

// DeleteUserByPhone godoc
// @Summary Delete a user account by phone number
// @Description Completely removes a user and all associated data (profile, skills, experience, education, applications, sessions, vectors)
// @Tags account-deletion
// @Param phone path string true "Phone number"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/by-phone/{phone} [delete]
func (h *AccountDeletionHandler) DeleteUserByPhone(c *gin.Context) {
	phone := c.Param("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "phone number is required"})
		return
	}

	if err := h.service.DeleteUserByPhone(c.Request.Context(), phone); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete user"})
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteHRByPhone godoc
// @Summary Delete an HR account by phone number
// @Description Removes an HR account and nullifies hr_id on their vacancies (vacancies are preserved)
// @Tags account-deletion
// @Param phone path string true "Phone number"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /hrs/by-phone/{phone} [delete]
func (h *AccountDeletionHandler) DeleteHRByPhone(c *gin.Context) {
	phone := c.Param("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "phone number is required"})
		return
	}

	if err := h.service.DeleteHRByPhone(c.Request.Context(), phone); err != nil {
		if errors.Is(err, domain.ErrCompanyHRNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "hr not found"})
			return
		}
		log.Printf("delete hr by phone %s: %v", phone, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete hr"})
		return
	}

	c.Status(http.StatusNoContent)
}
