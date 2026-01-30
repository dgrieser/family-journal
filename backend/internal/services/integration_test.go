package services

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/user/family-journal/internal/models"
	"github.com/user/family-journal/internal/repository"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&models.User{}, &models.Person{}, &models.Post{}, &models.Comment{}, &models.Hashtag{}, &models.Attachment{})
	assert.NoError(t, err)

	// Since we are using many2many, we might need to manually create the join tables or GORM handles it.
	// GORM handles it during AutoMigrate if tags are correct.

	return db
}

func TestRegistrationAndLogin(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo)

	// Test Register
	user, err := authService.Register("test@example.com", "password123", models.RoleUser)
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, models.RoleUser, user.Role)

	// Test Login
	loggedInUser, err := authService.Login("test@example.com", "password123")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, loggedInUser.ID)

	// Test Login Fail
	_, err = authService.Login("test@example.com", "wrongpassword")
	assert.Error(t, err)
}

func TestPostCreationWithHashtagsAndMentions(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	personRepo := repository.NewPersonRepository(db)
	postRepo := repository.NewPostRepository(db)

	authService := NewAuthService(userRepo)
	postService := NewPostService(postRepo, personRepo)

	user, _ := authService.Register("test@example.com", "password", models.RoleUser)

	text := "Care for @Child1 today. He is doing well #care #health"
	post, err := postService.CreatePost(user.ID, time.Now(), text)
	assert.NoError(t, err)
	assert.Len(t, post.Hashtags, 2)
	assert.Len(t, post.Mentions, 1)
	assert.Equal(t, "child1", post.Mentions[0].Name) // Wait, I didn't lowercase name in parseText but let's check.
}

func TestFiltering(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	personRepo := repository.NewPersonRepository(db)
	postRepo := repository.NewPostRepository(db)
	postService := NewPostService(postRepo, personRepo)

	user, _ := userRepo.FindByEmail("test@example.com") // setupTestDB doesn't persist across tests
	if user == nil {
		user = &models.User{Email: "test@example.com"}
		userRepo.Create(user)
	}

	today := time.Now()
	postService.CreatePost(user.ID, today, "Post 1 #tag1 @person1")
	postService.CreatePost(user.ID, today, "Post 2 #tag2 @person2")

	// Filter by hashtag
	posts, _ := postService.GetPosts(nil, []string{"tag1"}, nil, "")
	assert.Len(t, posts, 1)
	assert.Contains(t, posts[0].Text, "Post 1")

	// Filter by person
	posts, _ = postService.GetPosts(nil, nil, []string{"person2"}, "")
	assert.Len(t, posts, 1)
	assert.Contains(t, posts[0].Text, "Post 2")

	// Filter by search
	posts, _ = postService.GetPosts(nil, nil, nil, "Post 1")
	assert.Len(t, posts, 1)
}
