package services

import "familyjournal/backend/internal/models"

func (s *Service) CreateAttachment(att *models.Attachment) error {
	return s.Attachments.CreateAttachment(att)
}

func (s *Service) GetAttachmentForUser(scope AccessScope, name string) (*models.Attachment, error) {
	return s.Attachments.GetAttachmentByName(name, scope.OwnerFilter())
}
