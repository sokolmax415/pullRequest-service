package types

import (
	"encoding/json"
	"net/http"
)

type TeamRequestDTO struct {
	TeamName string          `json:"team_name"`
	Members  []TeamMemberDTO `json:"members"`
}

type TeamMemberDTO struct {
	UserID   string `json:"user_id"`
	UserName string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamResponseDTO struct {
	Team TeamRequestDTO `json:"team"`
}

func ParseTeamRequestDTO(r *http.Request) (*TeamRequestDTO, error) {
	var req TeamRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &req, nil
}
