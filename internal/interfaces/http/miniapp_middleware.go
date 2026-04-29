package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

// TelegramAuthMiddleware validates Telegram Mini App initData.
// The Authorization header should contain: "tma <initData>"
// See https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app
func TelegramAuthMiddleware(botToken string, userSvc *application.UserService, hrSvc *application.CompanyHRService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		log.Printf("DEBUG [UserBot]: Hit middleware. Authorization: '%s'", header)
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "missing authorization header"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "tma" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid authorization header, expected: tma <initData>"})
			return
		}

		initData := parts[1]

		telegramID, err := validateInitData(initData, botToken)
		if err != nil {
			log.Printf("miniapp auth failed: %v (initData length: %d)", err, len(initData))
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: fmt.Sprintf("invalid initData: %v", err)})
			return
		}

		tgIDStr := strconv.FormatInt(telegramID, 10)
		user, err := userSvc.GetByTelegramID(c.Request.Context(), tgIDStr)
		if err != nil {
			// Fallback: check if they are an HR
			hr, hrErr := hrSvc.GetByTelegramID(c.Request.Context(), tgIDStr)
			if hrErr == nil {
				// User is an HR, auto-create a candidate profile for them
				user, err = userSvc.CreateUser(c.Request.Context(), &domain.User{
					ID:         hr.ID,
					FirstName:  hr.FirstName,
					LastName:   hr.LastName,
					TelegramID: hr.TelegramID,
					Language:   hr.Language,
				})
				if err != nil {
					log.Printf("failed to auto-create user for HR %s in middleware: %v", hr.ID, err)
					c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "failed to initialize user profile"})
					return
				}
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
				return
			}
		}

		c.Set("user_id", user.ID.String())
		c.Set("tg_user_id", telegramID)
		c.Next()
	}
}

func validateInitData(initData, botToken string) (int64, error) {
	// Parse key=value pairs from the raw query string to preserve exact values.
	// We must use the URL-decoded values for the data-check-string.
	pairs := strings.Split(initData, "&")
	kv := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}
		decodedKey, err := url.QueryUnescape(k)
		if err != nil {
			decodedKey = k
		}
		decodedVal, err := url.QueryUnescape(v)
		if err != nil {
			decodedVal = v
		}
		kv[decodedKey] = decodedVal
	}

	hash, ok := kv["hash"]
	if !ok || hash == "" {
		return 0, fmt.Errorf("missing hash")
	}

	// Whitelisted hashes skip expiration check (for dev/testing tokens)
	whitelistedHashes := map[string]bool{
		"8be828f59fbe09521190c8344bf2af7461c6712331e0ac3adfede3b70458a143": true,
	}

	// Check auth_date is recent (within 24 hours)
	if !whitelistedHashes[hash] {
		if authDateStr, ok := kv["auth_date"]; ok {
			authDate, err := strconv.ParseInt(authDateStr, 10, 64)
			if err == nil {
				if time.Now().Unix()-authDate > 86400 {
					return 0, fmt.Errorf("initData expired")
				}
			}
		}
	}

	// Build the data-check-string: sort all key=value pairs (except hash), join with \n
	keys := make([]string, 0, len(kv))
	for k := range kv {
		if k == "hash" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var dataCheckParts []string
	for _, k := range keys {
		dataCheckParts = append(dataCheckParts, k+"="+kv[k])
	}
	dataCheckString := strings.Join(dataCheckParts, "\n")

	// secret_key = HMAC-SHA256("WebAppData", bot_token)
	secretKey := hmacSHA256([]byte("WebAppData"), []byte(botToken))

	// computed_hash = HMAC-SHA256(secret_key, data_check_string)
	computedHash := hex.EncodeToString(hmacSHA256(secretKey, []byte(dataCheckString)))

	if computedHash != hash {
		return 0, fmt.Errorf("hash mismatch")
	}

	// Extract user from initData
	userJSON, ok := kv["user"]
	if !ok || userJSON == "" {
		return 0, fmt.Errorf("missing user in initData")
	}

	var tgUser struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(userJSON), &tgUser); err != nil {
		return 0, fmt.Errorf("parse user JSON: %w", err)
	}

	if tgUser.ID == 0 {
		return 0, fmt.Errorf("invalid user ID")
	}

	return tgUser.ID, nil
}

// TelegramHRAuthMiddleware validates Telegram Mini App initData using the HR bot token.
func TelegramHRAuthMiddleware(hrBotToken string, hrSvc *application.CompanyHRService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			header = c.GetHeader("X-Telegram-Init-Data")
			if header != "" {
				header = "tma " + header
			}
		}
		log.Printf("DEBUG [HRBot]: Hit middleware. Authorization: '%s'", header)

		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "missing authorization header"})
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "tma" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid authorization header"})
			return
		}

		telegramID, err := validateInitData(parts[1], hrBotToken)
		if err != nil {
			log.Printf("hr miniapp auth failed: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: fmt.Sprintf("invalid initData: %v", err)})
			return
		}

		tgIDStr := strconv.FormatInt(telegramID, 10)
		hr, err := hrSvc.GetByTelegramID(c.Request.Context(), tgIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "hr not found"})
			return
		}

		c.Set("hr_id", hr.ID.String())
		c.Set("tg_user_id", telegramID)
		c.Next()
	}
}

// HRCombinedAuthMiddleware accepts both Telegram Mini App initData ("tma ...")
// and JWT Bearer tokens ("Bearer ..."). Both set "hr_id" in context.
func HRCombinedAuthMiddleware(jwtSecret, hrBotToken string, hrSvc *application.CompanyHRService) gin.HandlerFunc {
	tgMiddleware := TelegramHRAuthMiddleware(hrBotToken, hrSvc)
	jwtMiddleware := HRAuthMiddleware(jwtSecret)

	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			header = c.GetHeader("X-Telegram-Init-Data")
			if header != "" {
				header = "tma " + header
				c.Request.Header.Set("Authorization", header)
			}
		}
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "missing authorization header"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) < 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid authorization header"})
			return
		}

		switch parts[0] {
		case "tma":
			tgMiddleware(c)
		case "Bearer":
			jwtMiddleware(c)
		default:
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "unsupported authorization type"})
		}
	}
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
