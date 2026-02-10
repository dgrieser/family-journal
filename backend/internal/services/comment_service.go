package services

import "familyjournal/backend/internal/models"

func (s *Service) AddComment(comment *models.Comment) error {
	return s.Comments.CreateComment(comment)
}

func (s *Service) UpdateComment(scope AccessScope, comment *models.Comment) error {
	return s.Comments.UpdateComment(comment, scope.OwnerFilter())
}

func (s *Service) DeleteComment(scope AccessScope, commentID int64) error {
	return s.Comments.DeleteComment(commentID, scope.OwnerFilter())
}
