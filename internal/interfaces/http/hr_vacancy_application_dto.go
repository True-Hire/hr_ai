package http

// HRApplicantUserResponse contains user summary info for HR applicant listing.
type HRApplicantUserResponse struct {
	ID        string   `json:"id"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Phone     string   `json:"phone,omitempty"`
	Telegram  string   `json:"telegram,omitempty"`
	Skills    []string `json:"skills,omitempty"`
}

// HRApplicantResponse is a single application with user summary.
type HRApplicantResponse struct {
	ID          string                   `json:"id"`
	UserID      string                   `json:"user_id"`
	VacancyID   string                   `json:"vacancy_id"`
	Status      string                   `json:"status"`
	CoverLetter string                   `json:"cover_letter,omitempty"`
	SeenAt      *string                  `json:"seen_at"`
	CreatedAt   string                   `json:"created_at"`
	UpdatedAt   string                   `json:"updated_at"`
	User        *HRApplicantUserResponse `json:"user,omitempty"`
}

// HRApplicationListResponse is the paginated list of applicants with stats.
type HRApplicationListResponse struct {
	Applications []HRApplicantResponse `json:"applications"`
	Total        int64                 `json:"total"`
	Seen         int64                 `json:"seen"`
	Unseen       int64                 `json:"unseen"`
	Page         int32                 `json:"page"`
	PageSize     int32                 `json:"page_size"`
}

// HRApplicationStatsResponse contains application statistics for a vacancy.
type HRApplicationStatsResponse struct {
	Total  int64 `json:"total"`
	Seen   int64 `json:"seen"`
	Unseen int64 `json:"unseen"`
}

// HRUpdateApplicationStatusRequest is the request body for updating application status.
type HRUpdateApplicationStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
