package services

import (
	"regexp"
	"strings"
	"time"

	"github.com/user/family-journal/internal/models"
	"github.com/user/family-journal/internal/repository"
)

type PostService struct {
	postRepo   *repository.PostRepository
	personRepo *repository.PersonRepository
}

func NewPostService(postRepo *repository.PostRepository, personRepo *repository.PersonRepository) *PostService {
	return &PostService{
		postRepo:   postRepo,
		personRepo: personRepo,
	}
}

func (s *PostService) CreatePost(userID uint, date time.Time, text string) (*models.Post, error) {
	hashtags, mentions := s.parseText(text)

	post := &models.Post{
		UserID:   userID,
		Date:     date,
		Text:     text,
		Hashtags: hashtags,
		Mentions: mentions,
	}

	err := s.postRepo.Create(post)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostService) UpdatePost(postID uint, text string) (*models.Post, error) {
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return nil, err
	}

	hashtags, mentions := s.parseText(text)

	post.Text = text
	post.Hashtags = hashtags
	post.Mentions = mentions

	err = s.postRepo.Update(post)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostService) parseText(text string) ([]models.Hashtag, []models.Person) {
	hashtagRegex := regexp.MustCompile(`#(\w+)`)
	mentionRegex := regexp.MustCompile(`@(\w+)`)

	hashtagMatches := hashtagRegex.FindAllStringSubmatch(text, -1)
	mentionMatches := mentionRegex.FindAllStringSubmatch(text, -1)

	var hashtags []models.Hashtag
	hashtagMap := make(map[string]bool)
	for _, match := range hashtagMatches {
		name := strings.ToLower(match[1])
		if hashtagMap[name] {
			continue
		}
		hashtagMap[name] = true

		hashtag, err := s.postRepo.FindHashtagByName(name)
		if err != nil {
			// Create new hashtag
			hashtag = &models.Hashtag{Name: name}
			// We don't save it yet, GORM will save it via association
		}
		hashtags = append(hashtags, *hashtag)
	}

	var mentions []models.Person
	mentionMap := make(map[string]bool)
	for _, match := range mentionMatches {
		name := strings.ToLower(match[1])
		if mentionMap[name] {
			continue
		}
		mentionMap[name] = true

		person, err := s.personRepo.FindByName(name)
		if err != nil {
			// Create new person
			person = &models.Person{Name: name}
			// GORM will save it via association
		}
		mentions = append(mentions, *person)
	}

	return hashtags, mentions
}

func (s *PostService) GetPosts(date *time.Time, hashtags []string, persons []string, search string) ([]models.Post, error) {
	return s.postRepo.GetFiltered(date, hashtags, persons, search)
}

func (s *PostService) GetPost(id uint) (*models.Post, error) {
	return s.postRepo.FindByID(id)
}

func (s *PostService) DeletePost(id uint) error {
	return s.postRepo.Delete(id)
}

func (s *PostService) AddComment(postID uint, userID uint, text string) (*models.Comment, error) {
	comment := &models.Comment{
		PostID: postID,
		UserID: userID,
		Text:   text,
	}
	err := s.postRepo.AddComment(comment)
	return comment, err
}

func (s *PostService) DeleteComment(id uint) error {
	return s.postRepo.DeleteComment(id)
}

func (s *PostService) GetAllHashtags() ([]models.Hashtag, error) {
	return s.postRepo.GetAllHashtags()
}

func (s *PostService) AddAttachment(postID uint, fileName, fileType string, fileSize int64, storagePath string) (*models.Attachment, error) {
	attachment := &models.Attachment{
		PostID:      postID,
		FileName:    fileName,
		FileType:    fileType,
		FileSize:    fileSize,
		StoragePath: storagePath,
	}
	// Need a way to save attachment. I'll add it to post repo or generic.
	// For now, let's just use DB directly or add to post repo.
	err := s.postRepo.CreateAttachment(attachment)
	return attachment, err
}
