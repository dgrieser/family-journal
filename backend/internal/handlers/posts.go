package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"familyjournal/backend/internal/middleware"
	"familyjournal/backend/internal/models"
	"familyjournal/backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type PostsHandler struct {
	Service     *services.Service
	Store       *session.Store
	UploadDir   string
	MaxUploadMB int64
}

type postRequest struct {
	Date     string  `json:"date"`
	Text     string  `json:"text"`
	Category *string `json:"category"`
	Mood     *string `json:"mood"`
}

func (h *PostsHandler) List(c *fiber.Ctx) error {
	userID, _, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	dateParam := c.Query("date")
	if dateParam == "" {
		dateParam = time.Now().Format("2006-01-02")
	}
	date, err := time.Parse("2006-01-02", dateParam)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid date")
	}
	tags := splitQueryList(c.Query("hashtags"))
	persons := splitQueryList(c.Query("persons"))
	search := c.Query("search")
	posts, err := h.Service.ListPosts(userID, date, tags, persons, search)
	if err != nil {
		log.Printf("list posts error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list posts")
	}
	return c.JSON(posts)
}

func splitQueryList(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	var cleaned []string
	for _, part := range parts {
		trim := strings.TrimSpace(part)
		if trim != "" {
			cleaned = append(cleaned, trim)
		}
	}
	return cleaned
}

func (h *PostsHandler) Create(c *fiber.Ctx) error {
	userID, _, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	var req postRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid date")
	}
	post := &models.Post{
		UserID:   userID,
		Date:     date,
		Text:     req.Text,
		Category: req.Category,
		Mood:     req.Mood,
	}
	if err := h.Service.CreateOrUpdatePost(userID, post); err != nil {
		log.Printf("create post error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to save post")
	}
	return c.Status(fiber.StatusCreated).JSON(post)
}

func (h *PostsHandler) Update(c *fiber.Ctx) error {
	userID, _, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req postRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid date")
	}
	post := &models.Post{
		ID:       int64(id),
		UserID:   userID,
		Date:     date,
		Text:     req.Text,
		Category: req.Category,
		Mood:     req.Mood,
	}
	if err := h.Service.CreateOrUpdatePost(userID, post); err != nil {
		log.Printf("update post error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update post")
	}
	return c.JSON(post)
}

func (h *PostsHandler) Get(c *fiber.Ctx) error {
	userID, _, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	post, err := h.Service.GetPost(userID, int64(id))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "not found")
	}
	return c.JSON(post)
}

func (h *PostsHandler) Delete(c *fiber.Ctx) error {
	userID, _, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	if err := h.Service.DeletePost(userID, int64(id)); err != nil {
		log.Printf("delete post error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete post")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *PostsHandler) AddComment(c *fiber.Ctx) error {
	userID, _, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	postID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	if _, err := h.Service.GetPost(userID, int64(postID)); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "post not found")
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	comment := &models.Comment{PostID: int64(postID), UserID: userID, Text: req.Text}
	if err := h.Service.AddComment(comment); err != nil {
		log.Printf("add comment error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to add comment")
	}
	return c.Status(fiber.StatusCreated).JSON(comment)
}

func (h *PostsHandler) UpdateComment(c *fiber.Ctx) error {
	userID, _, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	commentID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	comment := &models.Comment{ID: int64(commentID), UserID: userID, Text: req.Text}
	if err := h.Service.UpdateComment(comment); err != nil {
		log.Printf("update comment error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update comment")
	}
	return c.JSON(comment)
}

func (h *PostsHandler) DeleteComment(c *fiber.Ctx) error {
	userID, _, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	commentID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	if err := h.Service.DeleteComment(userID, int64(commentID)); err != nil {
		log.Printf("delete comment error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete comment")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *PostsHandler) ListHashtags(c *fiber.Ctx) error {
	userID, _, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	tags, err := h.Service.ListHashtags(userID)
	if err != nil {
		log.Printf("list hashtags error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list hashtags")
	}
	return c.JSON(tags)
}

func (h *PostsHandler) UploadAttachment(c *fiber.Ctx) error {
	userID, _, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	postID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	if _, err := h.Service.GetPost(userID, int64(postID)); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "post not found")
	}
	form, err := c.MultipartForm()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid form")
	}
	files := form.File["files"]
	if len(files) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "no files")
	}
	var saved []models.Attachment
	for _, file := range files {
		if file.Size > h.MaxUploadMB*1024*1024 {
			return fiber.NewError(fiber.StatusBadRequest, "file too large")
		}
		contentType, err := detectFileType(file)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		if !isAllowedType(contentType) {
			return fiber.NewError(fiber.StatusBadRequest, "invalid file type")
		}
		fileName, err := uniqueFileName(file.Filename, contentType)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to generate filename")
		}
		path := filepath.Join(h.UploadDir, fileName)
		if err := c.SaveFile(file, path); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "save failed")
		}
		attachment := models.Attachment{
			PostID:   int64(postID),
			FileName: fileName,
			FileType: contentType,
			FileSize: file.Size,
			URL:      "/uploads/" + fileName,
		}
		if err := h.Service.CreateAttachment(&attachment); err != nil {
			log.Printf("create attachment error: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "failed to save attachment")
		}
		saved = append(saved, attachment)
	}
	return c.Status(fiber.StatusCreated).JSON(saved)
}

func isAllowedType(contentType string) bool {
	allowed := []string{"image/jpeg", "image/png", "application/pdf"}
	for _, t := range allowed {
		if t == contentType {
			return true
		}
	}
	return false
}

func detectFileType(file *multipart.FileHeader) (string, error) {
	reader, err := file.Open()
	if err != nil {
		return "", errors.New("unable to read file")
	}
	defer reader.Close()

	buffer := make([]byte, 512)
	n, err := reader.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", errors.New("unable to read file")
	}
	if n == 0 {
		return "", errors.New("empty file")
	}
	return http.DetectContentType(buffer[:n]), nil
}

func uniqueFileName(originalName, contentType string) (string, error) {
	base := filepath.Base(originalName)
	ext := filepath.Ext(base)
	if ext == "" {
		if derived, err := extensionForType(contentType); err == nil {
			ext = derived
		}
	}
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes) + ext, nil
}

func extensionForType(contentType string) (string, error) {
	extensions, err := mime.ExtensionsByType(contentType)
	if err != nil || len(extensions) == 0 {
		switch contentType {
		case "image/jpeg":
			return ".jpg", nil
		case "image/png":
			return ".png", nil
		case "application/pdf":
			return ".pdf", nil
		default:
			return "", errors.New("unsupported content type")
		}
	}
	return extensions[0], nil
}
