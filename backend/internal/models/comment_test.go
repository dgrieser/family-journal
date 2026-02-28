package models

import (
	"encoding/json"
	"testing"
)

func TestCommentHydrateUser(t *testing.T) {
	comment := Comment{UserID: 42, AuthorEmail: "author@example.com"}

	comment.HydrateUser()

	if comment.User.ID != 42 {
		t.Fatalf("expected User.ID 42, got %d", comment.User.ID)
	}
	if comment.User.Email != "author@example.com" {
		t.Fatalf("expected User.Email author@example.com, got %q", comment.User.Email)
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

	var payload map[string]any
	if err := json.Unmarshal(b, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if _, ok := payload["user_id"]; ok {
		t.Fatalf("expected user_id to be omitted from json, got payload: %s", string(b))
	}
	if _, ok := payload["author_email"]; ok {
		t.Fatalf("expected author_email to be omitted from json, got payload: %s", string(b))
	}

	userRaw, ok := payload["user"]
	if !ok {
		t.Fatalf("expected nested user field in json, got payload: %s", string(b))
	}
	user, ok := userRaw.(map[string]any)
	if !ok {
		t.Fatalf("expected user to be an object, got %T", userRaw)
	}
	if user["id"] != float64(3) {
		t.Fatalf("expected user.id 3, got %v", user["id"])
	}
	if user["email"] != "author@example.com" {
		t.Fatalf("expected user.email author@example.com, got %v", user["email"])
	}
}
