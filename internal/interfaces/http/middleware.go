package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "missing authorization header"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid authorization header"})
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token claims"})
			return
		}

		if claims["type"] != "access" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token type"})
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token subject"})
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

// JWTMiddleware parses JWT and sets user_id or hr_id + role in context.
// Does NOT enforce any specific role — use CasbinMiddleware for authorization.
func JWTMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "missing authorization header"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid authorization header"})
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token claims"})
			return
		}

		if claims["type"] != "access" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token type"})
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token subject"})
			return
		}

		role, _ := claims["role"].(string)
		if role == "hr" {
			c.Set("hr_id", sub)
			c.Set("role", "hr")
		} else {
			c.Set("user_id", sub)
			c.Set("role", "user")
		}

		c.Next()
	}
}

// CasbinMiddleware checks (role, resource, action) against Casbin policies.
func CasbinMiddleware(enforcer *casbin.Enforcer, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, ErrorResponse{Error: "no role in context"})
			return
		}

		allowed, err := enforcer.Enforce(role, resource, action)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: "authorization check failed"})
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, ErrorResponse{Error: "access denied"})
			return
		}

		c.Next()
	}
}

func HRAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "missing authorization header"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid authorization header"})
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token claims"})
			return
		}

		if claims["type"] != "access" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token type"})
			return
		}

		if claims["role"] != "hr" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token role"})
			return
		}

		hrID, ok := claims["sub"].(string)
		if !ok || hrID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token subject"})
			return
		}

		c.Set("hr_id", hrID)
		c.Set("role", "hr")
		c.Next()
	}
}
