package repository

import (
	"strings"
	"time"

	"github.com/user/family-journal/internal/models"
	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(post *models.Post) error {
	return r.db.Create(post).Error
}

func (r *PostRepository) Update(post *models.Post) error {
	return r.db.Save(post).Error
}

func (r *PostRepository) Delete(id uint) error {
	return r.db.Delete(&models.Post{}, id).Error
}

func (r *PostRepository) FindByID(id uint) (*models.Post, error) {
	var post models.Post
	err := r.db.Preload("User").Preload("Hashtags").Preload("Mentions").Preload("Attachments").Preload("Comments.User").First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) FindHashtagsByNames(names []string) ([]models.Hashtag, error) {
	var hashtags []models.Hashtag
	err := r.db.Where("name IN ?", names).Find(&hashtags).Error
	return hashtags, err
}

func (r *PostRepository) CreateHashtag(hashtag *models.Hashtag) error {
	return r.db.Create(hashtag).Error
}

func (r *PostRepository) GetFiltered(userID uint, date *time.Time, hashtags []string, persons []string, search string) ([]models.Post, error) {
	query := r.db.Preload("User").Preload("Hashtags").Preload("Mentions").Preload("Attachments").Preload("Comments.User")

	query = query.Where("posts.user_id = ?", userID)

	if date != nil {
		query = query.Where("date = ?", date.Format("2006-01-02"))
	}

	if len(hashtags) > 0 {
		query = query.Joins("JOIN post_hashtags ph ON ph.post_id = posts.id").
			Joins("JOIN hashtags h ON h.id = ph.hashtag_id").
			Where("h.name IN ?", hashtags)
	}

	if len(persons) > 0 {
		query = query.Joins("JOIN mentions m ON m.post_id = posts.id").
			Joins("JOIN persons p ON p.id = m.person_id").
			Where("p.name IN ?", persons)
	}

	if search != "" {
		// Escape wildcards for LIKE query to treat them as literal characters
		escapedSearch := strings.ReplaceAll(search, "\\", "\\\\")
		escapedSearch = strings.ReplaceAll(escapedSearch, "%", "\\%")
		escapedSearch = strings.ReplaceAll(escapedSearch, "_", "\\_")
		query = query.Where("text LIKE ?", "%"+escapedSearch+"%")
	}

	var posts []models.Post
	err := query.Order("posts.created_at DESC").Distinct().Find(&posts).Error
	return posts, err
}

func (r *PostRepository) AddComment(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

func (r *PostRepository) FindCommentByID(id uint) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *PostRepository) DeleteComment(id uint) error {
	return r.db.Delete(&models.Comment{}, id).Error
}

func (r *PostRepository) CreateAttachment(attachment *models.Attachment) error {
	return r.db.Create(attachment).Error
}

func (r *PostRepository) FindAttachmentByID(id uint) (*models.Attachment, error) {
	var attachment models.Attachment
	err := r.db.First(&attachment, id).Error
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

func (r *PostRepository) DeleteAttachment(id uint) error {
	return r.db.Delete(&models.Attachment{}, id).Error
}

func (r *PostRepository) GetAllHashtags() ([]models.Hashtag, error) {
	var hashtags []models.Hashtag
	err := r.db.Find(&hashtags).Error
	return hashtags, err
}
