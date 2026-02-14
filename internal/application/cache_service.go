package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	redisclient "github.com/ruziba3vich/hr-ai/internal/infrastructure/redis"
)

const (
	companyTTL       = 10 * time.Minute
	countryTTL       = 1 * time.Hour
	countriesListTTL = 1 * time.Hour
)

type CacheService struct {
	redis *redisclient.Client
}

func NewCacheService(redis *redisclient.Client) *CacheService {
	return &CacheService{redis: redis}
}

// Company

func (c *CacheService) GetCompany(ctx context.Context, id uuid.UUID) (*CompanyWithTexts, bool) {
	var result CompanyWithTexts
	found, err := c.redis.Get(ctx, companyKey(id), &result)
	if err != nil {
		return nil, false
	}
	return &result, found
}

func (c *CacheService) SetCompany(ctx context.Context, id uuid.UUID, data *CompanyWithTexts) {
	_ = c.redis.Set(ctx, companyKey(id), data, companyTTL)
}

func (c *CacheService) InvalidateCompany(ctx context.Context, id uuid.UUID) {
	_ = c.redis.Delete(ctx, companyKey(id))
}

// Country

func (c *CacheService) GetCountry(ctx context.Context, id uuid.UUID) (*CountryWithTexts, bool) {
	var result CountryWithTexts
	found, err := c.redis.Get(ctx, countryKey(id), &result)
	if err != nil {
		return nil, false
	}
	return &result, found
}

func (c *CacheService) SetCountry(ctx context.Context, id uuid.UUID, data *CountryWithTexts) {
	_ = c.redis.Set(ctx, countryKey(id), data, countryTTL)
}

func (c *CacheService) InvalidateCountry(ctx context.Context, id uuid.UUID) {
	_ = c.redis.Delete(ctx, countryKey(id), "countries:all")
}

// Countries list

func (c *CacheService) GetCountriesList(ctx context.Context) ([]CountryWithTexts, bool) {
	var result []CountryWithTexts
	found, err := c.redis.Get(ctx, "countries:all", &result)
	if err != nil {
		return nil, false
	}
	return result, found
}

func (c *CacheService) SetCountriesList(ctx context.Context, data []CountryWithTexts) {
	_ = c.redis.Set(ctx, "countries:all", data, countriesListTTL)
}

func (c *CacheService) InvalidateCountriesList(ctx context.Context) {
	_ = c.redis.Delete(ctx, "countries:all")
}

// Key builders

func companyKey(id uuid.UUID) string {
	return fmt.Sprintf("company:%s", id.String())
}

func countryKey(id uuid.UUID) string {
	return fmt.Sprintf("country:%s", id.String())
}
