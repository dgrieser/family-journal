package services

import (
	"errors"
	"regexp"
	"time"

	"familyjournal/backend/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInactiveUser       = errors.New("user is inactive")
	ErrInvalidPassword    = errors.New("password must be between 6 and 72 characters")
)

var hashtagRegex = regexp.MustCompile(`#([\pL\d_]+)`)
var mentionRegex = regexp.MustCompile(`@([\pL\d_]+)`)

type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id int64) (*models.User, error)
	UpdateUserProfile(id int64, email string) error
	UpdateUserPassword(id int64, passwordHash string) error
	ListUsers() ([]models.User, error)
	UpdateUserRole(id int64, role string) error
	UpdateUserActive(id int64, active bool) error
}

type PersonRepository interface {
	CreatePerson(person *models.Person) error
	UpdatePerson(person *models.Person, ownerFilter *int64) error
	DeletePerson(id int64, ownerFilter *int64) error
	ListPersons(ownerFilter *int64) ([]models.Person, error)
	FindOrCreatePerson(userID int64, name string) (*models.Person, error)
}

type HashtagRepository interface {
	ListHashtagsByUser(userID int64) ([]models.Hashtag, error)
	FindOrCreateHashtag(name string) (*models.Hashtag, error)
	ListTagsForPosts(postIDs []int64) (map[int64][]models.Hashtag, error)
}

type PostRepository interface {
	DeletePost(id int64, ownerFilter *int64) error
	GetPost(id int64, ownerFilter *int64) (*models.Post, error)
	ListPosts(ownerFilter *int64, date time.Time, hashtags, persons []string, search string) ([]models.Post, error)
	ListPersonsForPosts(postIDs []int64) (map[int64][]models.Person, error)
	ListCommentsForPosts(postIDs []int64) (map[int64][]models.Comment, error)
	ListAttachmentsForPosts(postIDs []int64) (map[int64][]models.Attachment, error)
	SavePostWithRelations(ownerID int64, ownerFilter *int64, post *models.Post, tagNames, personNames []string) error
}

type CommentRepository interface {
	CreateComment(comment *models.Comment) error
	UpdateComment(comment *models.Comment, ownerFilter *int64) error
	DeleteComment(id int64, ownerFilter *int64) error
}

type AttachmentRepository interface {
	CreateAttachment(att *models.Attachment) error
	GetAttachmentByName(name string, ownerFilter *int64) (*models.Attachment, error)
}

type Service struct {
	Users       UserRepository
	Persons     PersonRepository
	Hashtags    HashtagRepository
	Posts       PostRepository
	Comments    CommentRepository
	Attachments AttachmentRepository
}

func ensureSlice[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}

func New(users UserRepository, persons PersonRepository, hashtags HashtagRepository, posts PostRepository, comments CommentRepository, attachments AttachmentRepository) *Service {
	return &Service{
		Users:       users,
		Persons:     persons,
		Hashtags:    hashtags,
		Posts:       posts,
		Comments:    comments,
		Attachments: attachments,
	}
}
