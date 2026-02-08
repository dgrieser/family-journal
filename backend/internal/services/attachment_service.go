package services

import "familyjournal/backend/internal/models"

func (s *Service) CreateAttachment(att *models.Attachment) error {
	return s.Attachments.CreateAttachment(att)
}

func (s *Service) GetAttachmentForUser(userID int64, name string) (*models.Attachment, error) {
	return s.Attachments.GetAttachmentByName(userID, name)
}
