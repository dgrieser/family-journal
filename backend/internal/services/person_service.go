package services

import "familyjournal/backend/internal/models"

func (s *Service) ListPersons(scope AccessScope, pagination PaginationParams) (PaginatedResponse[models.Person], error) {
	totalItems, err := s.Persons.CountPersons(scope.OwnerFilter())
	if err != nil {
		return PaginatedResponse[models.Person]{}, err
	}
	persons, err := s.Persons.ListPersons(scope.OwnerFilter(), pagination.PageSize, pagination.Offset())
	if err != nil {
		return PaginatedResponse[models.Person]{}, err
	}
	return NewPaginatedResponse(ensureSlice(persons), totalItems, pagination), nil
}

func (s *Service) CreatePerson(userID int64, name string, description *string) (*models.Person, error) {
	person := &models.Person{Name: name, Description: description, CreatedBy: userID}
	if err := s.Persons.CreatePerson(person); err != nil {
		return nil, err
	}
	return person, nil
}

func (s *Service) UpdatePerson(scope AccessScope, person *models.Person) error {
	return s.Persons.UpdatePerson(person, scope.OwnerFilter())
}

func (s *Service) DeletePerson(scope AccessScope, personID int64) error {
	return s.Persons.DeletePerson(personID, scope.OwnerFilter())
}
