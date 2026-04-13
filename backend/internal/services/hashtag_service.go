package services

import (
	"strings"

	"familyjournal/backend/internal/models"
)

func (s *Service) ListAllHashtags() ([]models.Hashtag, error) {
	tags, err := s.Hashtags.ListAllHashtags()
	if err != nil {
		return nil, err
	}
	return ensureSlice(tags), nil
}

func (s *Service) CreateHashtag(userID int64, name string) (*models.Hashtag, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	tag := &models.Hashtag{Name: name, CreatedByUserID: &userID}
	if err := s.Hashtags.CreateHashtag(tag); err != nil {
		return nil, err
	}
	return tag, nil
}

func (s *Service) UpdateHashtag(scope AccessScope, tag *models.Hashtag) error {
	tag.Name = strings.ToLower(strings.TrimSpace(tag.Name))
	return s.Hashtags.UpdateHashtag(tag, scope.OwnerFilter())
}

func (s *Service) DeleteHashtag(scope AccessScope, id int64) error {
	return s.Hashtags.DeleteHashtag(id, scope.OwnerFilter())
}
