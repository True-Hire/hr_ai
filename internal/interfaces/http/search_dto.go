package http

type SearchUsersResponse struct {
	Users []SearchResultItem `json:"users"`
	Query string             `json:"query"`
	Total int                `json:"total"`
}

type SearchResultItem struct {
	User  UserResponse `json:"user"`
	Score float64      `json:"score"`
}
