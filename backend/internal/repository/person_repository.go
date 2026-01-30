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

func (r *PersonRepository) FindByName(name string) (*models.Person, error) {
	var person models.Person
	err := r.db.Where("name = ?", name).First(&person).Error
	if err != nil {
		return nil, err
	}
	return &person, nil
}

func (r *PersonRepository) GetAll() ([]models.Person, error) {
	var persons []models.Person
	err := r.db.Find(&persons).Error
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
