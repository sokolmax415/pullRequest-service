package entity

import "fmt"

type Team struct {
	TeamName string
	Members  []TeamMember
}

type TeamMember struct {
	UserID   string
	UserName string
	IsActive bool
}

func (t *Team) Validate() error {
	if t.TeamName == "" {
		return fmt.Errorf("%w: empty team_name", ErrInvalidRequest)
	}

	for _, m := range t.Members {
		if m.UserID == "" || m.UserName == "" {
			return fmt.Errorf("%w: empty user data", ErrInvalidRequest)
		}
	}
	return nil
}
