package models

import (
	"encoding/json"
	"testing"
)

func TestCommentHydrateUser(t *testing.T) {
	comment := Comment{UserID: 42, AuthorEmail: "author@example.com"}

	comment.HydrateUser()

	if comment.User.ID != 42 {
		t.Errorf("expected User.ID 42, got %d", comment.User.ID)
	}
	if comment.User.Email != "author@example.com" {
		t.Errorf("expected User.Email author@example.com, got %q", comment.User.Email)
	}
}

func TestCommentJSONOmitsLegacyAuthorFields(t *testing.T) {
	comment := Comment{
		ID:          1,
		PostID:      2,
		UserID:      3,
		Text:        "hello",
		AuthorEmail: "author@example.com",
	}
	comment.HydrateUser()

	b, err := json.Marshal(comment)
	if err != nil {
		t.Fatalf("marshal comment: %v", err)
	}

	var result struct {
		User struct {
			ID    int64  `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
		UserID      *int64  `json:"user_id"`
		AuthorEmail *string `json:"author_email"`
	}
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if result.UserID != nil {
		t.Errorf("expected user_id to be omitted from json, but got: %d", *result.UserID)
	}
	if result.AuthorEmail != nil {
		t.Errorf("expected author_email to be omitted from json, but got: %q", *result.AuthorEmail)
	}
	if result.User.ID != 3 {
		t.Errorf("expected user.id 3, got %d", result.User.ID)
	}
	if result.User.Email != "author@example.com" {
		t.Errorf("expected user.email \"author@example.com\", got %q", result.User.Email)
	}
}
