package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	postToReturn   *models.Post
	deletedPostID  int64
	deletedUserID  int64
	ownerFilterNil bool
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

func (f *fakeRepo) UpdateUserProfile(id int64, email string) error               { return nil }
func (f *fakeRepo) UpdateUserPassword(id int64, passwordHash string) error       { return nil }
func (f *fakeRepo) ListUsers() ([]models.User, error)                            { return nil, nil }
func (f *fakeRepo) UpdateUserRole(id int64, role string) error                   { return nil }
func (f *fakeRepo) UpdateUserActive(id int64, active bool) error                 { return nil }
func (f *fakeRepo) CreatePerson(person *models.Person) error                     { return nil }
func (f *fakeRepo) UpdatePerson(person *models.Person, ownerFilter *int64) error { return nil }
func (f *fakeRepo) DeletePerson(id int64, ownerFilter *int64) error              { return nil }
func (f *fakeRepo) ListPersons(ownerFilter *int64) ([]models.Person, error)      { return nil, nil }
func (f *fakeRepo) FindOrCreatePerson(userID int64, name string) (*models.Person, error) {
	f.personsCreated = append(f.personsCreated, name)
	return &models.Person{ID: int64(len(f.personsCreated)), Name: name}, nil
}
func (f *fakeRepo) ListHashtags(ownerFilter *int64) ([]models.Hashtag, error) {
	return nil, nil
}
func (f *fakeRepo) FindOrCreateHashtag(name string) (*models.Hashtag, error) {
	f.tagsCreated = append(f.tagsCreated, name)
	return &models.Hashtag{ID: int64(len(f.tagsCreated)), Name: name}, nil
}
func (f *fakeRepo) CreatePost(post *models.Post) error { post.ID = 1; return nil }
func (f *fakeRepo) UpdatePost(post *models.Post) error { return nil }
func (f *fakeRepo) DeletePost(id int64, ownerFilter *int64) error {
	f.deletedPostID = id
	f.ownerFilterNil = ownerFilter == nil
	if ownerFilter != nil {
		f.deletedUserID = *ownerFilter
	}
	return nil
}
func (f *fakeRepo) GetPost(id int64, ownerFilter *int64) (*models.Post, error) {
	if f.postToReturn != nil {
		return f.postToReturn, nil
	}
	return &models.Post{}, nil
}
func (f *fakeRepo) ListPosts(ownerFilter *int64, date time.Time, hashtags, persons []string, search string) ([]models.Post, error) {
	f.listPostsArgs.date = date
	f.listPostsArgs.hashtags = hashtags
	f.listPostsArgs.persons = persons
	f.listPostsArgs.search = search
	return nil, nil
}
func (f *fakeRepo) ReplacePostTags(postID int64, tags []models.Hashtag) error       { return nil }
func (f *fakeRepo) ReplacePostMentions(postID int64, persons []models.Person) error { return nil }
func (f *fakeRepo) ListPostComments(postID int64) ([]models.Comment, error)         { return nil, nil }
func (f *fakeRepo) CreateComment(comment *models.Comment) error                     { return nil }
func (f *fakeRepo) UpdateComment(comment *models.Comment, ownerFilter *int64) error { return nil }
func (f *fakeRepo) DeleteComment(id int64, ownerFilter *int64) error                { return nil }
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
	result := map[int64][]models.Attachment{}
	if f.postToReturn == nil {
		return result, nil
	}
	for _, postID := range postIDs {
		if postID == f.postToReturn.ID {
			result[postID] = f.postToReturn.Attachments
		}
	}
	return result, nil
}
func (f *fakeRepo) SavePostWithRelations(ownerID int64, ownerFilter *int64, post *models.Post, tagNames, personNames []string) error {
	f.tagsCreated = append(f.tagsCreated, tagNames...)
	f.personsCreated = append(f.personsCreated, personNames...)
	if post.ID == 0 {
		post.ID = 1
	}
	return nil
}
func (f *fakeRepo) GetAttachmentByID(id int64, ownerFilter *int64) (*models.Attachment, error) {
	return nil, sql.ErrNoRows
}

func TestJSONErrorHandlerReturnsErrorObject(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: handlers.JSONErrorHandler})
	app.Get("/bad", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadRequest, "cannot parse json")
	})

	req := httptest.NewRequest(http.MethodGet, "/bad", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	var got map[string]string
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("expected json error response, got %q: %v", string(body), err)
	}
	if got["error"] != "cannot parse json" {
		t.Fatalf("expected error message, got %#v", got)
	}
}

func TestJSONErrorHandlerMasksInternalErrors(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: handlers.JSONErrorHandler})
	app.Get("/explode", func(c *fiber.Ctx) error {
		return errors.New("db connection failed: password=super-secret")
	})

	req := httptest.NewRequest(http.MethodGet, "/explode", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	var got map[string]string
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("expected json error response, got %q: %v", string(body), err)
	}
	if got["error"] != "internal server error" {
		t.Fatalf("expected masked error message, got %#v", got)
	}
}

func TestJSONErrorHandlerMasksFiberServerErrors(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: handlers.JSONErrorHandler})
	app.Get("/panic", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusInternalServerError, "sql query failed: select * from users")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	var got map[string]string
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("expected json error response, got %q: %v", string(body), err)
	}
	if got["error"] != "internal server error" {
		t.Fatalf("expected masked fiber server error message, got %#v", got)
	}
}

func TestJSONErrorHandlerUsesStatusTextForClientErrorsWithoutMessage(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: handlers.JSONErrorHandler})
	app.Get("/missing", func(c *fiber.Ctx) error {
		return &fiber.Error{Code: fiber.StatusNotFound}
	})

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	var got map[string]string
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("expected json error response, got %q: %v", string(body), err)
	}
	if got["error"] != "Not Found" {
		t.Fatalf("expected status text error message, got %#v", got)
	}
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
	if err := service.CreateOrUpdatePost(services.NewAccessScope(1, models.RoleUser), post); err != nil {
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
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		t.Fatalf("read list posts response: %v", readErr)
	}
	if string(body) != "[]" {
		t.Fatalf("expected empty array response, got %s", string(body))
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

func TestAdminDeletePostUsesUnscopedAccess(t *testing.T) {
	repo := newFakeRepo()
	service := services.New(repo, repo, repo, repo, repo, repo)

	err := service.DeletePost(services.NewAccessScope(99, models.RoleAdmin), 42)
	if err != nil {
		t.Fatalf("delete post: %v", err)
	}
	if !repo.ownerFilterNil {
		t.Fatalf("expected admin delete to skip owner filter")
	}
}

func TestDeletePostRemovesAttachmentFiles(t *testing.T) {
	repo := newFakeRepo()
	repo.postToReturn = &models.Post{
		ID: 1,
		Attachments: []models.Attachment{
			{FileName: "test-file.jpg"},
		},
	}

	service := services.New(repo, repo, repo, repo, repo, repo)
	store := session.New()
	uploadDir := t.TempDir()
	filePath := filepath.Join(uploadDir, "test-file.jpg")
	if err := os.WriteFile(filePath, []byte("content"), 0o600); err != nil {
		t.Fatalf("write attachment file: %v", err)
	}

	app := fiber.New()
	postsHandler := &handlers.PostsHandler{Service: service, Store: store, UploadDir: uploadDir}
	app.Use(func(c *fiber.Ctx) error {
		sess, _ := store.Get(c)
		sess.Set("user_id", int64(1))
		sess.Set("role", "user")
		_ = sess.Save()
		return c.Next()
	})
	app.Delete("/posts/:id", postsHandler.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/posts/1", nil)
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete post failed: %v", err)
	}

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("expected attachment file to be removed")
	}
	if repo.deletedPostID != 1 || repo.deletedUserID != 1 {
		t.Fatalf("expected delete to be called with post and user ids")
	}
}

func TestDeletePostAttemptsAllAttachmentDeletesOnError(t *testing.T) {
	repo := newFakeRepo()
	repo.postToReturn = &models.Post{
		ID: 1,
		Attachments: []models.Attachment{
			{FileName: "blocked"},
			{FileName: "still-delete.jpg"},
		},
	}

	service := services.New(repo, repo, repo, repo, repo, repo)
	store := session.New()
	uploadDir := t.TempDir()

	blockedDir := filepath.Join(uploadDir, "blocked")
	if err := os.Mkdir(blockedDir, 0o700); err != nil {
		t.Fatalf("create blocked dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(blockedDir, "child.txt"), []byte("x"), 0o600); err != nil {
		t.Fatalf("create nested file: %v", err)
	}

	goodFile := filepath.Join(uploadDir, "still-delete.jpg")
	if err := os.WriteFile(goodFile, []byte("content"), 0o600); err != nil {
		t.Fatalf("write attachment file: %v", err)
	}

	app := fiber.New()
	postsHandler := &handlers.PostsHandler{Service: service, Store: store, UploadDir: uploadDir}
	app.Use(func(c *fiber.Ctx) error {
		sess, _ := store.Get(c)
		sess.Set("user_id", int64(1))
		sess.Set("role", "user")
		_ = sess.Save()
		return c.Next()
	})
	app.Delete("/posts/:id", postsHandler.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/posts/1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("delete post request failed: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected internal server error, got %d", resp.StatusCode)
	}

	if _, err := os.Stat(goodFile); !os.IsNotExist(err) {
		t.Fatalf("expected later attachment file to still be deleted")
	}
	if repo.deletedPostID != 0 || repo.deletedUserID != 0 {
		t.Fatalf("did not expect post delete repository call when attachment cleanup fails")
	}
}

func TestDeletePostUsesBaseFileNameForAttachmentCleanup(t *testing.T) {
	repo := newFakeRepo()
	repo.postToReturn = &models.Post{
		ID: 1,
		Attachments: []models.Attachment{
			{FileName: "../outside.jpg"},
		},
	}

	service := services.New(repo, repo, repo, repo, repo, repo)
	store := session.New()
	uploadDir := t.TempDir()
	parentDir := filepath.Dir(uploadDir)

	outsideFile := filepath.Join(parentDir, "outside.jpg")
	if err := os.WriteFile(outsideFile, []byte("outside"), 0o600); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	insideFile := filepath.Join(uploadDir, "outside.jpg")
	if err := os.WriteFile(insideFile, []byte("inside"), 0o600); err != nil {
		t.Fatalf("write inside file: %v", err)
	}

	app := fiber.New()
	postsHandler := &handlers.PostsHandler{Service: service, Store: store, UploadDir: uploadDir}
	app.Use(func(c *fiber.Ctx) error {
		sess, _ := store.Get(c)
		sess.Set("user_id", int64(1))
		sess.Set("role", "user")
		_ = sess.Save()
		return c.Next()
	})
	app.Delete("/posts/:id", postsHandler.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/posts/1", nil)
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete post failed: %v", err)
	}

	if _, err := os.Stat(insideFile); !os.IsNotExist(err) {
		t.Fatalf("expected in-upload attachment file to be deleted")
	}
	if _, err := os.Stat(outsideFile); err != nil {
		t.Fatalf("expected file outside upload dir to remain untouched: %v", err)
	}
}

func TestServiceNormalizesNilSlices(t *testing.T) {
	repo := newFakeRepo()
	service := services.New(repo, repo, repo, repo, repo, repo)

	users, err := service.ListUsers()
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if users == nil {
		t.Fatalf("expected users slice to be non-nil")
	}

	persons, err := service.ListPersons(services.NewAccessScope(1, models.RoleUser))
	if err != nil {
		t.Fatalf("list persons: %v", err)
	}
	if persons == nil {
		t.Fatalf("expected persons slice to be non-nil")
	}

	tags, err := service.ListHashtags(services.NewAccessScope(1, models.RoleUser))
	if err != nil {
		t.Fatalf("list hashtags: %v", err)
	}
	if tags == nil {
		t.Fatalf("expected hashtags slice to be non-nil")
	}

	posts, err := service.ListPosts(services.NewAccessScope(1, models.RoleUser), time.Now(), nil, nil, "")
	if err != nil {
		t.Fatalf("list posts: %v", err)
	}
	if posts == nil {
		t.Fatalf("expected posts slice to be non-nil")
	}

	post, err := service.GetPost(services.NewAccessScope(1, models.RoleUser), 1)
	if err != nil {
		t.Fatalf("get post: %v", err)
	}
	if post.Hashtags == nil || post.Persons == nil || post.Comments == nil || post.Attachments == nil {
		t.Fatalf("expected hydrated post collections to be non-nil")
	}
}
