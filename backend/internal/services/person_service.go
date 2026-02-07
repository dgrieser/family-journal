package services

import "familyjournal/backend/internal/models"

func (s *Service) ListPersons(userID int64) ([]models.Person, error) {
	persons, err := s.Persons.ListPersons(userID)
	if err != nil {
		return nil, err
	}
	return ensureSlice(persons), nil
}

func (s *Service) CreatePerson(userID int64, name string, description *string) (*models.Person, error) {
	person := &models.Person{Name: name, Description: description, CreatedBy: userID}
	if err := s.Persons.CreatePerson(person); err != nil {
		return nil, err
	}
	return person, nil
}

func (s *Service) UpdatePerson(userID int64, person *models.Person) error {
	person.CreatedBy = userID
	return s.Persons.UpdatePerson(person)
}

func (s *Service) DeletePerson(userID, personID int64) error {
	return s.Persons.DeletePerson(personID, userID)
}
