package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ruziba3vich/hr-ai/internal/application"
)

// TelegramAuthMiddleware validates Telegram Mini App initData.
// The Authorization header should contain: "tma <initData>"
// See https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app
func TelegramAuthMiddleware(botToken string, userSvc *application.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
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
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: fmt.Sprintf("invalid initData: %v", err)})
			return
		}

		tgIDStr := strconv.FormatInt(telegramID, 10)
		user, err := userSvc.GetByTelegramID(c.Request.Context(), tgIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
			return
		}

		c.Set("user_id", user.ID.String())
		c.Set("tg_user_id", telegramID)
		c.Next()
	}
}

func validateInitData(initData, botToken string) (int64, error) {
	// Parse the initData query string
	values, err := url.ParseQuery(initData)
	if err != nil {
		return 0, fmt.Errorf("parse initData: %w", err)
	}

	hash := values.Get("hash")
	if hash == "" {
		return 0, fmt.Errorf("missing hash")
	}

	// Check auth_date is recent (within 24 hours)
	authDateStr := values.Get("auth_date")
	if authDateStr != "" {
		authDate, err := strconv.ParseInt(authDateStr, 10, 64)
		if err == nil {
			if time.Now().Unix()-authDate > 86400 {
				return 0, fmt.Errorf("initData expired")
			}
		}
	}

	// Build the data-check-string: sort all key=value pairs (except hash), join with \n
	values.Del("hash")
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var dataCheckParts []string
	for _, k := range keys {
		dataCheckParts = append(dataCheckParts, k+"="+values.Get(k))
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
	userJSON := values.Get("user")
	if userJSON == "" {
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

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
