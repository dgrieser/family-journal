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

type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id int64) (*models.User, error)
	UpdateUserProfile(id int64, email string) error
	ListUsers() ([]models.User, error)
	UpdateUserRole(id int64, role string) error
	UpdateUserActive(id int64, active bool) error
}

type PersonRepository interface {
	CreatePerson(person *models.Person) error
	UpdatePerson(person *models.Person) error
	DeletePerson(id, userID int64) error
	ListPersons(userID int64) ([]models.Person, error)
	FindOrCreatePerson(userID int64, name string) (*models.Person, error)
}

type HashtagRepository interface {
	ListHashtagsByUser(userID int64) ([]models.Hashtag, error)
	FindOrCreateHashtag(name string) (*models.Hashtag, error)
	ListTagsForPosts(postIDs []int64) (map[int64][]models.Hashtag, error)
}

type PostRepository interface {
	CreatePost(post *models.Post) error
	UpdatePost(post *models.Post) error
	DeletePost(id, userID int64) error
	GetPost(id, userID int64) (*models.Post, error)
	ListPosts(userID int64, date time.Time, hashtags, persons []string, search string) ([]models.Post, error)
	ReplacePostTags(postID int64, tags []models.Hashtag) error
	ReplacePostMentions(postID int64, persons []models.Person) error
	ListPersonsForPosts(postIDs []int64) (map[int64][]models.Person, error)
	ListCommentsForPosts(postIDs []int64) (map[int64][]models.Comment, error)
	ListAttachmentsForPosts(postIDs []int64) (map[int64][]models.Attachment, error)
	SavePostWithRelations(userID int64, post *models.Post, tagNames, personNames []string) error
}

type CommentRepository interface {
	CreateComment(comment *models.Comment) error
	UpdateComment(comment *models.Comment) error
	DeleteComment(id, userID int64) error
}

type AttachmentRepository interface {
	CreateAttachment(att *models.Attachment) error
	GetAttachmentByName(userID int64, name string) (*models.Attachment, error)
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

func (s *Service) Register(email, password string) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &models.User{
		Email:    email,
		Password: string(hash),
		Role:     models.RoleUser,
		Active:   true,
	}
	if err := s.Users.CreateUser(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) Authenticate(email, password string) (*models.User, error) {
	user, err := s.Users.GetUserByEmail(email)
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
	tags := s.ParseHashtags(post.Text)
	persons := s.ParseMentions(post.Text)
	return s.Posts.SavePostWithRelations(userID, post, tags, persons)
}

func (s *Service) ListPosts(userID int64, date time.Time, hashtags, persons []string, search string) ([]models.Post, error) {
	posts, err := s.Posts.ListPosts(userID, date, hashtags, persons, search)
	if err != nil {
		return nil, err
	}
	return s.hydratePosts(ensureSlice(posts))
}

func (s *Service) GetPost(userID, postID int64) (*models.Post, error) {
	post, err := s.Posts.GetPost(postID, userID)
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

func (s *Service) ListHashtags(userID int64) ([]models.Hashtag, error) {
	tags, err := s.Hashtags.ListHashtagsByUser(userID)
	if err != nil {
		return nil, err
	}
	return ensureSlice(tags), nil
}

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

func (s *Service) GetUserByID(userID int64) (*models.User, error) {
	return s.Users.GetUserByID(userID)
}

func (s *Service) UpdateUserProfile(userID int64, email string) error {
	return s.Users.UpdateUserProfile(userID, email)
}

func (s *Service) ListUsers() ([]models.User, error) {
	users, err := s.Users.ListUsers()
	if err != nil {
		return nil, err
	}
	return ensureSlice(users), nil
}

func (s *Service) UpdateUserRole(userID int64, role string) error {
	return s.Users.UpdateUserRole(userID, role)
}

func (s *Service) UpdateUserActive(userID int64, active bool) error {
	return s.Users.UpdateUserActive(userID, active)
}

func (s *Service) DeletePost(userID, postID int64) error {
	return s.Posts.DeletePost(postID, userID)
}

func (s *Service) AddComment(comment *models.Comment) error {
	return s.Comments.CreateComment(comment)
}

func (s *Service) UpdateComment(comment *models.Comment) error {
	return s.Comments.UpdateComment(comment)
}

func (s *Service) DeleteComment(userID, commentID int64) error {
	return s.Comments.DeleteComment(commentID, userID)
}

func (s *Service) CreateAttachment(att *models.Attachment) error {
	return s.Attachments.CreateAttachment(att)
}

func (s *Service) GetAttachmentForUser(userID int64, name string) (*models.Attachment, error) {
	return s.Attachments.GetAttachmentByName(userID, name)
}

func (s *Service) hydratePosts(posts []models.Post) ([]models.Post, error) {
	if len(posts) == 0 {
		return ensureSlice(posts), nil
	}
	ids := make([]int64, 0, len(posts))
	for _, post := range posts {
		ids = append(ids, post.ID)
	}
	tagsByPost, err := s.Hashtags.ListTagsForPosts(ids)
	if err != nil {
		return nil, err
	}
	personsByPost, err := s.Posts.ListPersonsForPosts(ids)
	if err != nil {
		return nil, err
	}
	commentsByPost, err := s.Posts.ListCommentsForPosts(ids)
	if err != nil {
		return nil, err
	}
	attachmentsByPost, err := s.Posts.ListAttachmentsForPosts(ids)
	if err != nil {
		return nil, err
	}
	for i := range posts {
		posts[i].Hashtags = ensureSlice(tagsByPost[posts[i].ID])
		posts[i].Persons = ensureSlice(personsByPost[posts[i].ID])
		posts[i].Comments = ensureSlice(commentsByPost[posts[i].ID])
		posts[i].Attachments = ensureSlice(attachmentsByPost[posts[i].ID])
	}
	return posts, nil
}
