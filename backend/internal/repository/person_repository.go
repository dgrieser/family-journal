package repository

import (
	"github.com/user/family-journal/internal/models"
	"gorm.io/gorm"
)

type PersonRepository struct {
	db *gorm.DB
}

func NewPersonRepository(db *gorm.DB) *PersonRepository {
	return &PersonRepository{db: db}
}

func (r *PersonRepository) Create(person *models.Person) error {
	return r.db.Create(person).Error
}

func (r *PersonRepository) FindByNames(userID uint, names []string) ([]models.Person, error) {
	var persons []models.Person
	err := r.db.Where("created_by_user_id = ? AND name IN ?", userID, names).Find(&persons).Error
	return persons, err
}

func (r *PersonRepository) GetAll(userID uint) ([]models.Person, error) {
	var persons []models.Person
	err := r.db.Where("created_by_user_id = ?", userID).Find(&persons).Error
	return persons, err
}

func (r *PersonRepository) Delete(id uint) error {
	return r.db.Delete(&models.Person{}, id).Error
}

func (r *PersonRepository) Update(person *models.Person) error {
	return r.db.Save(person).Error
}

func (r *PersonRepository) FindByID(id uint) (*models.Person, error) {
	var person models.Person
	err := r.db.First(&person, id).Error
	if err != nil {
		return nil, err
	}
	return &person, nil
}
