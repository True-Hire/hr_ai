package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CandidateSearchProfile struct {
	UserID                uuid.UUID
	PrimaryRole           string
	RoleFamily            string
	Seniority             string
	TotalExperienceMonths int
	HighestEducationLevel string

	Skills          []string
	Industries      []string
	ProjectDomains  []string
	CompanyNames    []string
	KnownLanguages  []string
	EducationFields []string
	Universities    []string

	LocationCity      string
	LocationCountry   string
	WillingToRelocate bool

	// Role scores
	BackendScore  float64
	FrontendScore float64
	MobileScore   float64
	DataScore     float64
	QAScore       float64
	PMScore       float64
	DevOpsScore   float64
	DesignScore   float64

	// Capability scores
	DevOpsSupportScore       float64
	ClientCommunicationScore float64
	ProjectManagementScore   float64
	OwnershipScore           float64
	LeadershipScore          float64
	MentoringScore           float64
	StartupAdaptabilityScore float64

	// Market strength scores
	CompanyPrestigeScore        float64
	EngineeringEnvironmentScore float64
	InternshipQualityScore      float64
	EducationQualityScore       float64
	CompetitionScore            float64
	OpenSourceScore             float64
	GrowthTrajectoryScore       float64
	ProjectComplexityScore      float64

	// Aggregated
	OverallStrengthScore  float64
	BackendStrengthScore  float64
	FrontendStrengthScore float64
	DataStrengthScore     float64

	SearchText string

	ScoringFactors map[string]interface{}
	ParsedEntities map[string]interface{}

	UpdatedAt time.Time
}

type CandidateSearchProfileRepository interface {
	Upsert(ctx context.Context, profile *CandidateSearchProfile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*CandidateSearchProfile, error)
	Delete(ctx context.Context, userID uuid.UUID) error
	SearchPool(ctx context.Context, query string, roleFamily string, seniority string, locationCity string, locationCountry string, poolSize int) ([]CandidateSearchProfile, error)
	SearchPoolBySkills(ctx context.Context, skills []string, roleFamily string, locationCity string, poolSize int) ([]CandidateSearchProfile, error)
}
