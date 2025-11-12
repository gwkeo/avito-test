package dto

type TeamMember struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}
