package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"familyjournal/backend/internal/handlers"
	"familyjournal/backend/internal/models"
	"familyjournal/backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type fakeRepo struct {
	users          map[string]*models.User
	tagsCreated    []string
	personsCreated []string
	listPostsArgs  struct {
		date     time.Time
		hashtags []string
		persons  []string
		search   string
	}
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{users: map[string]*models.User{}}
}

func (f *fakeRepo) CreateUser(user *models.User) error {
	user.ID = int64(len(f.users) + 1)
	f.users[user.Email] = user
	return nil
}

func (f *fakeRepo) GetUserByEmail(email string) (*models.User, error) {
	user, ok := f.users[email]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return user, nil
}

func (f *fakeRepo) GetUserByID(id int64) (*models.User, error) {
	for _, user := range f.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, fiber.ErrNotFound
}

func (f *fakeRepo) UpdateUserProfile(id int64, email string) error    { return nil }
func (f *fakeRepo) ListUsers() ([]models.User, error)                 { return nil, nil }
func (f *fakeRepo) UpdateUserRole(id int64, role string) error        { return nil }
func (f *fakeRepo) UpdateUserActive(id int64, active bool) error      { return nil }
func (f *fakeRepo) CreatePerson(person *models.Person) error          { return nil }
func (f *fakeRepo) UpdatePerson(person *models.Person) error          { return nil }
func (f *fakeRepo) DeletePerson(id, userID int64) error               { return nil }
func (f *fakeRepo) ListPersons(userID int64) ([]models.Person, error) { return nil, nil }
func (f *fakeRepo) FindOrCreatePerson(userID int64, name string) (*models.Person, error) {
	f.personsCreated = append(f.personsCreated, name)
	return &models.Person{ID: int64(len(f.personsCreated)), Name: name}, nil
}
func (f *fakeRepo) ListHashtags() ([]models.Hashtag, error) { return nil, nil }
func (f *fakeRepo) ListHashtagsByUser(userID int64) ([]models.Hashtag, error) {
	return nil, nil
}
func (f *fakeRepo) FindOrCreateHashtag(name string) (*models.Hashtag, error) {
	f.tagsCreated = append(f.tagsCreated, name)
	return &models.Hashtag{ID: int64(len(f.tagsCreated)), Name: name}, nil
}
func (f *fakeRepo) CreatePost(post *models.Post) error             { post.ID = 1; return nil }
func (f *fakeRepo) UpdatePost(post *models.Post) error             { return nil }
func (f *fakeRepo) DeletePost(id, userID int64) error              { return nil }
func (f *fakeRepo) GetPost(id, userID int64) (*models.Post, error) { return &models.Post{}, nil }
func (f *fakeRepo) ListPosts(userID int64, date time.Time, hashtags, persons []string, search string) ([]models.Post, error) {
	f.listPostsArgs.date = date
	f.listPostsArgs.hashtags = hashtags
	f.listPostsArgs.persons = persons
	f.listPostsArgs.search = search
	return []models.Post{}, nil
}
func (f *fakeRepo) ReplacePostTags(postID int64, tags []models.Hashtag) error       { return nil }
func (f *fakeRepo) ReplacePostMentions(postID int64, persons []models.Person) error { return nil }
func (f *fakeRepo) ListPostComments(postID int64) ([]models.Comment, error)         { return nil, nil }
func (f *fakeRepo) CreateComment(comment *models.Comment) error                     { return nil }
func (f *fakeRepo) UpdateComment(comment *models.Comment) error                     { return nil }
func (f *fakeRepo) DeleteComment(id, userID int64) error                            { return nil }
func (f *fakeRepo) ListPostTags(postID int64) ([]models.Hashtag, error)             { return nil, nil }
func (f *fakeRepo) ListPostPersons(postID int64) ([]models.Person, error)           { return nil, nil }
func (f *fakeRepo) ListPostAttachments(postID int64) ([]models.Attachment, error)   { return nil, nil }
func (f *fakeRepo) CreateAttachment(att *models.Attachment) error                   { return nil }
func (f *fakeRepo) ListTagsForPosts(postIDs []int64) (map[int64][]models.Hashtag, error) {
	return map[int64][]models.Hashtag{}, nil
}
func (f *fakeRepo) ListPersonsForPosts(postIDs []int64) (map[int64][]models.Person, error) {
	return map[int64][]models.Person{}, nil
}
func (f *fakeRepo) ListCommentsForPosts(postIDs []int64) (map[int64][]models.Comment, error) {
	return map[int64][]models.Comment{}, nil
}
func (f *fakeRepo) ListAttachmentsForPosts(postIDs []int64) (map[int64][]models.Attachment, error) {
	return map[int64][]models.Attachment{}, nil
}
func (f *fakeRepo) SavePostWithRelations(userID int64, post *models.Post, tagNames, personNames []string) error {
	f.tagsCreated = append(f.tagsCreated, tagNames...)
	f.personsCreated = append(f.personsCreated, personNames...)
	if post.ID == 0 {
		post.ID = 1
	}
	return nil
}
func (f *fakeRepo) GetAttachmentByName(userID int64, name string) (*models.Attachment, error) {
	return nil, sql.ErrNoRows
}

func TestRegisterLoginSession(t *testing.T) {
	repo := newFakeRepo()
	service := services.New(repo, repo, repo, repo, repo, repo)
	store := session.New()
	app := fiber.New()
	h := &handlers.AuthHandler{Service: service, Store: store}
	app.Post("/register", h.Register)
	app.Post("/login", h.Login)
	app.Get("/profile", h.Profile)

	payload, _ := json.Marshal(map[string]string{"email": "test@example.com", "password": "secret"})
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("register failed: %v", err)
	}
	registerCookie := resp.Header.Get("Set-Cookie")
	if registerCookie != "" {
		t.Fatalf("did not expect session cookie on register")
	}

	profileReq := httptest.NewRequest(http.MethodGet, "/profile", nil)
	profileResp, err := app.Test(profileReq)
	if err != nil || profileResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized profile right after register: %v", err)
	}

	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed: %v", err)
	}
	cookie := resp.Header.Get("Set-Cookie")
	if cookie == "" {
		t.Fatalf("expected session cookie")
	}

	profileReq = httptest.NewRequest(http.MethodGet, "/profile", nil)
	profileReq.Header.Set("Cookie", cookie)
	profileResp, err = app.Test(profileReq)
	if err != nil || profileResp.StatusCode != http.StatusOK {
		t.Fatalf("profile failed right after login: %v", err)
	}
}

func TestCreatePostParsesTagsAndPersons(t *testing.T) {
	repo := newFakeRepo()
	service := services.New(repo, repo, repo, repo, repo, repo)
	post := &models.Post{UserID: 1, Date: time.Now(), Text: "Today #Care with @Lena"}
	if err := service.CreateOrUpdatePost(1, post); err != nil {
		t.Fatalf("create post: %v", err)
	}
	if len(repo.tagsCreated) != 1 || repo.tagsCreated[0] != "care" {
		t.Fatalf("expected hashtag creation")
	}
	if len(repo.personsCreated) != 1 || repo.personsCreated[0] != "Lena" {
		t.Fatalf("expected person creation")
	}
}

func TestListPostsFilters(t *testing.T) {
	repo := newFakeRepo()
	service := services.New(repo, repo, repo, repo, repo, repo)
	store := session.New()
	app := fiber.New()
	postsHandler := &handlers.PostsHandler{Service: service, Store: store}
	app.Use(func(c *fiber.Ctx) error {
		sess, _ := store.Get(c)
		sess.Set("user_id", int64(1))
		sess.Set("role", "user")
		_ = sess.Save()
		return c.Next()
	})
	app.Get("/posts", postsHandler.List)

	date := time.Now().Format("2006-01-02")
	req := httptest.NewRequest(http.MethodGet, "/posts?date="+date+"&hashtags=care,food&persons=Lena&search=note", nil)
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("list posts failed: %v", err)
	}
	if repo.listPostsArgs.search != "note" {
		t.Fatalf("expected search param")
	}
	if len(repo.listPostsArgs.hashtags) != 2 {
		t.Fatalf("expected hashtags")
	}
	if len(repo.listPostsArgs.persons) != 1 {
		t.Fatalf("expected persons")
	}
}
