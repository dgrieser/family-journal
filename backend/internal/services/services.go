package services

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"familyjournal/backend/internal/models"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInactiveUser       = errors.New("user is inactive")
)

var hashtagRegex = regexp.MustCompile(`#([\pL\d_]+)`)
var mentionRegex = regexp.MustCompile(`@([\pL\d_]+)`)

type Repository interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id int64) (*models.User, error)
	UpdateUserProfile(id int64, email string) error
	ListUsers() ([]models.User, error)
	UpdateUserRole(id int64, role string) error
	UpdateUserActive(id int64, active bool) error
	CreatePerson(person *models.Person) error
	UpdatePerson(person *models.Person) error
	DeletePerson(id, userID int64) error
	ListPersons(userID int64) ([]models.Person, error)
	FindOrCreatePerson(userID int64, name string) (*models.Person, error)
	ListHashtags() ([]models.Hashtag, error)
	FindOrCreateHashtag(name string) (*models.Hashtag, error)
	CreatePost(post *models.Post) error
	UpdatePost(post *models.Post) error
	DeletePost(id, userID int64) error
	GetPost(id, userID int64) (*models.Post, error)
	ListPosts(userID int64, date time.Time, hashtags, persons []string, search string) ([]models.Post, error)
	ReplacePostTags(postID int64, tags []models.Hashtag) error
	ReplacePostMentions(postID int64, persons []models.Person) error
	ListPostComments(postID int64) ([]models.Comment, error)
	CreateComment(comment *models.Comment) error
	UpdateComment(comment *models.Comment) error
	DeleteComment(id, userID int64) error
	ListPostTags(postID int64) ([]models.Hashtag, error)
	ListPostPersons(postID int64) ([]models.Person, error)
	ListPostAttachments(postID int64) ([]models.Attachment, error)
	CreateAttachment(att *models.Attachment) error
	ListTagsForPosts(postIDs []int64) (map[int64][]models.Hashtag, error)
	ListPersonsForPosts(postIDs []int64) (map[int64][]models.Person, error)
	ListCommentsForPosts(postIDs []int64) (map[int64][]models.Comment, error)
	ListAttachmentsForPosts(postIDs []int64) (map[int64][]models.Attachment, error)
}

type Service struct {
	Repo Repository
}

func New(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) Register(email, password string) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &models.User{
		Email:    email,
		Password: string(hash),
		Role:     "user",
		Active:   true,
	}
	if err := s.Repo.CreateUser(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) Authenticate(email, password string) (*models.User, error) {
	user, err := s.Repo.GetUserByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.Active {
		return nil, ErrInactiveUser
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (s *Service) ParseHashtags(text string) []string {
	matches := hashtagRegex.FindAllStringSubmatch(text, -1)
	unique := map[string]struct{}{}
	var tags []string
	for _, match := range matches {
		tag := strings.ToLower(match[1])
		if _, ok := unique[tag]; !ok {
			unique[tag] = struct{}{}
			tags = append(tags, tag)
		}
	}
	return tags
}

func (s *Service) ParseMentions(text string) []string {
	matches := mentionRegex.FindAllStringSubmatch(text, -1)
	unique := map[string]struct{}{}
	var names []string
	for _, match := range matches {
		name := match[1]
		if _, ok := unique[name]; !ok {
			unique[name] = struct{}{}
			names = append(names, name)
		}
	}
	return names
}

func (s *Service) CreateOrUpdatePost(userID int64, post *models.Post) error {
	var err error
	if post.ID == 0 {
		err = s.Repo.CreatePost(post)
	} else {
		err = s.Repo.UpdatePost(post)
	}
	if err != nil {
		return err
	}

	tags := s.ParseHashtags(post.Text)
	persons := s.ParseMentions(post.Text)
	var tagModels []models.Hashtag
	for _, tag := range tags {
		model, err := s.Repo.FindOrCreateHashtag(tag)
		if err != nil {
			return err
		}
		tagModels = append(tagModels, *model)
	}
	var personModels []models.Person
	for _, name := range persons {
		model, err := s.Repo.FindOrCreatePerson(userID, name)
		if err != nil {
			return err
		}
		personModels = append(personModels, *model)
	}
	if err := s.Repo.ReplacePostTags(post.ID, tagModels); err != nil {
		return err
	}
	if err := s.Repo.ReplacePostMentions(post.ID, personModels); err != nil {
		return err
	}
	return nil
}

func (s *Service) ListPosts(userID int64, date time.Time, hashtags, persons []string, search string) ([]models.Post, error) {
	posts, err := s.Repo.ListPosts(userID, date, hashtags, persons, search)
	if err != nil {
		return nil, err
	}
	return s.hydratePosts(posts)
}

func (s *Service) GetPost(userID, postID int64) (*models.Post, error) {
	post, err := s.Repo.GetPost(postID, userID)
	if err != nil {
		return nil, err
	}
	posts, err := s.hydratePosts([]models.Post{*post})
	if err != nil {
		return nil, err
	}
	if len(posts) == 0 {
		return nil, errors.New("post not found")
	}
	return &posts[0], nil
}

func (s *Service) ListHashtags() ([]models.Hashtag, error) {
	return s.Repo.ListHashtags()
}

func (s *Service) ListPersons(userID int64) ([]models.Person, error) {
	return s.Repo.ListPersons(userID)
}

func (s *Service) CreatePerson(userID int64, name string, description *string) (*models.Person, error) {
	person := &models.Person{Name: name, Description: description, CreatedBy: userID}
	if err := s.Repo.CreatePerson(person); err != nil {
		return nil, err
	}
	return person, nil
}

func (s *Service) UpdatePerson(userID int64, person *models.Person) error {
	person.CreatedBy = userID
	return s.Repo.UpdatePerson(person)
}

func (s *Service) DeletePerson(userID, personID int64) error {
	return s.Repo.DeletePerson(personID, userID)
}

func (s *Service) GetUserByID(userID int64) (*models.User, error) {
	return s.Repo.GetUserByID(userID)
}

func (s *Service) UpdateUserProfile(userID int64, email string) error {
	return s.Repo.UpdateUserProfile(userID, email)
}

func (s *Service) ListUsers() ([]models.User, error) {
	return s.Repo.ListUsers()
}

func (s *Service) UpdateUserRole(userID int64, role string) error {
	return s.Repo.UpdateUserRole(userID, role)
}

func (s *Service) UpdateUserActive(userID int64, active bool) error {
	return s.Repo.UpdateUserActive(userID, active)
}

func (s *Service) DeletePost(userID, postID int64) error {
	return s.Repo.DeletePost(postID, userID)
}

func (s *Service) AddComment(comment *models.Comment) error {
	return s.Repo.CreateComment(comment)
}

func (s *Service) UpdateComment(comment *models.Comment) error {
	return s.Repo.UpdateComment(comment)
}

func (s *Service) DeleteComment(userID, commentID int64) error {
	return s.Repo.DeleteComment(commentID, userID)
}

func (s *Service) CreateAttachment(att *models.Attachment) error {
	return s.Repo.CreateAttachment(att)
}

func (s *Service) hydratePosts(posts []models.Post) ([]models.Post, error) {
	if len(posts) == 0 {
		return posts, nil
	}
	ids := make([]int64, 0, len(posts))
	for _, post := range posts {
		ids = append(ids, post.ID)
	}
	tagsByPost, err := s.Repo.ListTagsForPosts(ids)
	if err != nil {
		return nil, err
	}
	personsByPost, err := s.Repo.ListPersonsForPosts(ids)
	if err != nil {
		return nil, err
	}
	commentsByPost, err := s.Repo.ListCommentsForPosts(ids)
	if err != nil {
		return nil, err
	}
	attachmentsByPost, err := s.Repo.ListAttachmentsForPosts(ids)
	if err != nil {
		return nil, err
	}
	for i := range posts {
		posts[i].Hashtags = tagsByPost[posts[i].ID]
		posts[i].Persons = personsByPost[posts[i].ID]
		posts[i].Comments = commentsByPost[posts[i].ID]
		posts[i].Attachments = attachmentsByPost[posts[i].ID]
	}
	return posts, nil
}
