package models

import "time"

type HashtagCreator struct {
	Email string `json:"email"`
}

type Hashtag struct {
	ID              int64           `db:"id" json:"id"`
	Name            string          `db:"name" json:"name"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	CreatedByUserID *int64          `db:"created_by_user_id" json:"created_by_user_id,omitempty"`
	CreatorEmail    string          `db:"creator_email" json:"-"`
	Creator         *HashtagCreator `json:"creator,omitempty"`
}

func (h *Hashtag) HydrateCreator() {
	if h.CreatorEmail != "" {
		h.Creator = &HashtagCreator{Email: h.CreatorEmail}
	}
}
