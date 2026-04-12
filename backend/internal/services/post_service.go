package services

import (
	"strings"
	"time"

	"familyjournal/backend/internal/models"
)

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

func (s *Service) CreateOrUpdatePost(scope AccessScope, post *models.Post) error {
	ownerID := scope.UserID
	if post.ID != 0 {
		existing, err := s.Posts.GetPost(post.ID, scope.OwnerFilter())
		if err != nil {
			return err
		}
		ownerID = existing.UserID
	}
	tags := s.ParseHashtags(post.Text)
	persons := s.ParseMentions(post.Text)
	return s.Posts.SavePostWithRelations(ownerID, scope.OwnerFilter(), post, tags, persons)
}

func (s *Service) ListPosts(scope AccessScope, date time.Time, hashtags, persons []string, search string, pagination PaginationParams) (PaginatedResponse[models.Post], error) {
	posts, totalItems, err := s.Posts.ListPostsPaginated(nil, date, hashtags, persons, search, pagination.PageSize, pagination.Offset())
	if err != nil {
		return PaginatedResponse[models.Post]{}, err
	}
	hydrated, err := s.hydratePosts(ensureSlice(posts))
	if err != nil {
		return PaginatedResponse[models.Post]{}, err
	}
	return NewPaginatedResponse(hydrated, totalItems, pagination), nil
}

func (s *Service) GetPost(scope AccessScope, postID int64) (*models.Post, error) {
	post, err := s.Posts.GetPost(postID, nil)
	if err != nil {
		return nil, err
	}
	posts, err := s.hydratePosts([]models.Post{*post})
	if err != nil {
		return nil, err
	}
	return &posts[0], nil
}

func (s *Service) DeletePost(scope AccessScope, postID int64) error {
	return s.Posts.DeletePost(postID, scope.OwnerFilter())
}

func (s *Service) ListHashtags(scope AccessScope) ([]models.Hashtag, error) {
	tags, err := s.Hashtags.ListHashtags(nil)
	if err != nil {
		return nil, err
	}
	return ensureSlice(tags), nil
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
