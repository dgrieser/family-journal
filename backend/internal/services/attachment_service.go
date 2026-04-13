package services

import "familyjournal/backend/internal/models"

func (s *Service) CreateAttachment(att *models.Attachment) error {
	return s.Attachments.CreateAttachment(att)
}

func (s *Service) GetAttachmentByIDForUser(scope AccessScope, id int64) (*models.Attachment, error) {
	return s.Attachments.GetAttachmentByID(id, nil)
}

func (s *Service) DeleteAttachmentByID(scope AccessScope, id int64) error {
	return s.Attachments.DeleteAttachmentByID(id, scope.OwnerFilter())
}
