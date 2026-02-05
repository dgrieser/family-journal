package services

import (
	"os"
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
	hashtags, mentions, err := s.parseText(text, userID)
	if err != nil {
		return nil, err
	}

	post := &models.Post{
		UserID:   userID,
		Date:     date,
		Text:     text,
		Hashtags: hashtags,
		Mentions: mentions,
	}

	err = s.postRepo.Create(post)
	if err != nil {
		return nil, err
	}

	return s.postRepo.FindByID(post.ID)
}

func (s *PostService) UpdatePost(postID uint, text string, date *time.Time) (*models.Post, error) {
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(text) != "" {
		hashtags, mentions, err := s.parseText(text, post.UserID)
		if err != nil {
			return nil, err
		}
		post.Text = text
		post.Hashtags = hashtags
		post.Mentions = mentions
	}
	if date != nil {
		post.Date = *date
	}

	err = s.postRepo.Update(post)
	if err != nil {
		return nil, err
	}

	return s.postRepo.FindByID(post.ID)
}

func (s *PostService) parseText(text string, userID uint) ([]models.Hashtag, []models.Person, error) {
	hashtagRegex := regexp.MustCompile(`#(\w+)`)
	mentionRegex := regexp.MustCompile(`@(\w+)`)

	hashtagMatches := hashtagRegex.FindAllStringSubmatch(text, -1)
	mentionMatches := mentionRegex.FindAllStringSubmatch(text, -1)

	var uniqueHashtagNames []string
	hashtagMap := make(map[string]bool)
	for _, match := range hashtagMatches {
		name := strings.ToLower(match[1])
		if !hashtagMap[name] {
			hashtagMap[name] = true
			uniqueHashtagNames = append(uniqueHashtagNames, name)
		}
	}

	var uniqueMentionNames []string
	mentionMap := make(map[string]bool)
	for _, match := range mentionMatches {
		name := strings.ToLower(match[1])
		if !mentionMap[name] {
			mentionMap[name] = true
			uniqueMentionNames = append(uniqueMentionNames, name)
		}
	}

	var hashtags []models.Hashtag
	if len(uniqueHashtagNames) > 0 {
		existingHashtags, err := s.postRepo.FindHashtagsByNames(uniqueHashtagNames)
		if err != nil {
			return nil, nil, err
		}
		existingMap := make(map[string]models.Hashtag)
		for _, h := range existingHashtags {
			existingMap[h.Name] = h
		}

		for _, name := range uniqueHashtagNames {
			if h, ok := existingMap[name]; ok {
				hashtags = append(hashtags, h)
			} else {
				hashtags = append(hashtags, models.Hashtag{Name: name})
			}
		}
	}

	var mentions []models.Person
	if len(uniqueMentionNames) > 0 {
		// Security: Only find persons created by THIS user to avoid information leak
		existingPersons, err := s.personRepo.FindByNames(userID, uniqueMentionNames)
		if err != nil {
			return nil, nil, err
		}
		existingMap := make(map[string]models.Person)
		for _, p := range existingPersons {
			existingMap[strings.ToLower(p.Name)] = p
		}

		for _, name := range uniqueMentionNames {
			if p, ok := existingMap[strings.ToLower(name)]; ok {
				mentions = append(mentions, p)
			} else {
				mentions = append(mentions, models.Person{
					Name:            name,
					CreatedByUserID: &userID,
				})
			}
		}
	}

	return hashtags, mentions, nil
}

func (s *PostService) GetPosts(userID uint, date *time.Time, hashtags []string, persons []string, search string) ([]models.Post, error) {
	return s.postRepo.GetFiltered(userID, date, hashtags, persons, search)
}

func (s *PostService) GetPost(id uint) (*models.Post, error) {
	return s.postRepo.FindByID(id)
}

func (s *PostService) DeletePost(id uint) error {
	post, err := s.postRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Delete physical files
	for _, a := range post.Attachments {
		_ = os.Remove(a.StoragePath)
	}

	return s.postRepo.Delete(id)
}

func (s *PostService) GetComment(id uint) (*models.Comment, error) {
	return s.postRepo.FindCommentByID(id)
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

func (s *PostService) GetAttachment(id uint) (*models.Attachment, error) {
	return s.postRepo.FindAttachmentByID(id)
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
