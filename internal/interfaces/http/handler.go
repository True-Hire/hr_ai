package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type UserHandler struct {
	service *application.UserService
}

func NewUserHandler(service *application.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	created, err := h.service.CreateUser(c.Request.Context(), req.ToDomain())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, toUserResponse(created))
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
		return
	}

	user, err := h.service.GetUser(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get user"})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *UserHandler) List(c *gin.Context) {
	page := parseQueryInt32(c, "page", 1)
	pageSize := parseQueryInt32(c, "page_size", 20)

	result, err := h.service.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list users"})
		return
	}

	resp := PaginatedUsersResponse{
		Users:    make([]UserResponse, 0, len(result.Users)),
		Total:    result.Total,
		Page:     page,
		PageSize: pageSize,
	}
	for _, u := range result.Users {
		resp.Users = append(resp.Users, toUserResponse(&u))
	}

	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	updated, err := h.service.UpdateUser(c.Request.Context(), req.ToDomain(id))
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(updated))
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
		return
	}

	if err := h.service.DeleteUser(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete user"})
		return
	}

	c.Status(http.StatusNoContent)
}

func parseQueryInt32(c *gin.Context, key string, defaultVal int32) int32 {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return int32(n)
}
