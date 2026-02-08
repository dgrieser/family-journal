package services

import "familyjournal/backend/internal/models"

func (s *Service) AddComment(comment *models.Comment) error {
	return s.Comments.CreateComment(comment)
}

func (s *Service) UpdateComment(comment *models.Comment) error {
	return s.Comments.UpdateComment(comment)
}

func (s *Service) DeleteComment(userID, commentID int64) error {
	return s.Comments.DeleteComment(commentID, userID)
}
