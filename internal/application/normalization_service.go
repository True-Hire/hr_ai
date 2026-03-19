package application

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type NormalizationService struct {
	repo  domain.NormalizationRuleRepository
	mu    sync.RWMutex
	cache map[string]map[string]string // category -> source_value -> normalized_value
}

func NewNormalizationService(repo domain.NormalizationRuleRepository) *NormalizationService {
	return &NormalizationService{
		repo:  repo,
		cache: make(map[string]map[string]string),
	}
}

// LoadCache loads all rules from DB into memory. Called once at startup.
func (s *NormalizationService) LoadCache(ctx context.Context) error {
	rules, err := s.repo.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("load normalization rules: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache = make(map[string]map[string]string)
	for _, r := range rules {
		if s.cache[r.Category] == nil {
			s.cache[r.Category] = make(map[string]string)
		}
		s.cache[r.Category][r.SourceValue] = r.NormalizedValue
	}
	return nil
}

// NormalizeRole - implements NormalizerInterface
// Iterates "role" category entries, checks if title contains source_value (like current hardcoded behavior)
func (s *NormalizationService) NormalizeRole(title string) string {
	lower := strings.ToLower(strings.TrimSpace(title))
	s.mu.RLock()
	defer s.mu.RUnlock()
	if m, ok := s.cache["role"]; ok {
		for source, normalized := range m {
			if strings.Contains(lower, source) {
				return normalized
			}
		}
	}
	return lower
}

// RoleFamily - implements NormalizerInterface
// Looks up "role_family" category. The source_value is a keyword that might appear in the role string.
func (s *NormalizationService) RoleFamily(role string) string {
	lower := strings.ToLower(strings.TrimSpace(role))
	s.mu.RLock()
	defer s.mu.RUnlock()
	if m, ok := s.cache["role_family"]; ok {
		for source, family := range m {
			if strings.Contains(lower, source) {
				return family
			}
		}
	}
	return "other"
}

// NormalizeSkill - implements NormalizerInterface
// Exact match on "skill" category
func (s *NormalizationService) NormalizeSkill(skill string) string {
	lower := strings.ToLower(strings.TrimSpace(skill))
	s.mu.RLock()
	defer s.mu.RUnlock()
	if m, ok := s.cache["skill"]; ok {
		if normalized, exists := m[lower]; exists {
			return normalized
		}
	}
	return lower
}

// NormalizeCompany - implements NormalizerInterface
// Iterates "company" entries, checks if name contains source_value
func (s *NormalizationService) NormalizeCompany(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	s.mu.RLock()
	defer s.mu.RUnlock()
	if m, ok := s.cache["company"]; ok {
		for source, normalized := range m {
			if strings.Contains(lower, source) {
				return normalized
			}
		}
	}
	return lower
}

// ExtractDomains - implements NormalizerInterface
// Checks "domain" category keywords against combined text
func (s *NormalizationService) ExtractDomains(texts ...string) []string {
	combined := strings.ToLower(strings.Join(texts, " "))
	s.mu.RLock()
	defer s.mu.RUnlock()

	seen := make(map[string]bool)
	var domains []string
	if m, ok := s.cache["domain"]; ok {
		for keyword, domainName := range m {
			if strings.Contains(combined, keyword) && !seen[domainName] {
				domains = append(domains, domainName)
				seen[domainName] = true
			}
		}
	}
	return domains
}

// DetermineSeniority is logic-based, not a lookup. Keep as pass-through.
func (s *NormalizationService) DetermineSeniority(totalMonths int, hasLeadership bool) string {
	switch {
	case totalMonths < 12:
		return "intern"
	case totalMonths < 30:
		return "junior"
	case totalMonths < 60:
		return "middle"
	default:
		if hasLeadership {
			return "lead"
		}
		return "senior"
	}
}

// --- CRUD methods for HTTP handler ---

func (s *NormalizationService) CreateRule(ctx context.Context, rule *domain.NormalizationRule) (*domain.NormalizationRule, error) {
	created, err := s.repo.Create(ctx, rule)
	if err != nil {
		return nil, err
	}
	s.addToCache(created.Category, created.SourceValue, created.NormalizedValue)
	return created, nil
}

func (s *NormalizationService) GetRule(ctx context.Context, id uuid.UUID) (*domain.NormalizationRule, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *NormalizationService) UpdateRule(ctx context.Context, rule *domain.NormalizationRule) (*domain.NormalizationRule, error) {
	updated, err := s.repo.Update(ctx, rule)
	if err != nil {
		return nil, err
	}
	// Reload cache to handle category/source changes
	_ = s.LoadCache(ctx)
	return updated, nil
}

func (s *NormalizationService) DeleteRule(ctx context.Context, id uuid.UUID) error {
	// Get the rule first to know what to remove from cache
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	// Reload cache
	_ = s.LoadCache(context.Background())
	return nil
}

type NormalizationRuleListResult struct {
	Rules []domain.NormalizationRule
	Total int64
}

func (s *NormalizationService) ListRules(ctx context.Context, category, query string, page, pageSize int) (*NormalizationRuleListResult, error) {
	offset := (page - 1) * pageSize
	rules, err := s.repo.Search(ctx, category, query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.repo.Count(ctx, category, query)
	if err != nil {
		return nil, err
	}
	return &NormalizationRuleListResult{Rules: rules, Total: total}, nil
}

// EnsureNormalized checks if a source_value exists for the given category.
// If not, inserts it with the provided normalized_value. Thread-safe.
func (s *NormalizationService) EnsureNormalized(ctx context.Context, category, sourceValue, normalizedValue string) {
	sourceValue = strings.ToLower(strings.TrimSpace(sourceValue))
	if sourceValue == "" || normalizedValue == "" {
		return
	}

	// Check cache first (fast path)
	s.mu.RLock()
	if m, ok := s.cache[category]; ok {
		if _, exists := m[sourceValue]; exists {
			s.mu.RUnlock()
			return
		}
	}
	s.mu.RUnlock()

	// Not in cache — upsert into DB
	_, _ = s.repo.Upsert(ctx, &domain.NormalizationRule{
		Category:        category,
		SourceValue:     sourceValue,
		NormalizedValue: normalizedValue,
		Metadata:        map[string]any{},
	})

	// Update cache
	s.addToCache(category, sourceValue, normalizedValue)
}

func (s *NormalizationService) addToCache(category, source, normalized string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cache[category] == nil {
		s.cache[category] = make(map[string]string)
	}
	s.cache[category][source] = normalized
}
