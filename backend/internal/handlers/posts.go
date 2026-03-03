package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
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
	Service      *services.Service
	Store        *session.Store
	UploadDir    string
	MaxUploadMB  int64
	AllowedTypes []string
}

type postRequest struct {
	Date string `json:"date"`
	Text string `json:"text"`
}

func ensureSlice[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}

func normalizePostCollections(post *models.Post) {
	post.Hashtags = ensureSlice(post.Hashtags)
	post.Persons = ensureSlice(post.Persons)
	post.Comments = ensureSlice(post.Comments)
	post.Attachments = ensureSlice(post.Attachments)
}

func (h *PostsHandler) List(c *fiber.Ctx) error {
	userID, role, err := middleware.GetSessionUser(c, h.Store)
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
	pagination, err := parsePagination(c)
	if err != nil {
		return err
	}
	scope := services.NewAccessScope(userID, role)
	posts, err := h.Service.ListPosts(scope, date, tags, persons, search, pagination)
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
	userID, role, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	var req postRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	if strings.TrimSpace(req.Text) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "text is required")
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid date")
	}
	post := &models.Post{
		UserID: userID,
		Date:   date,
		Text:   req.Text,
	}
	scope := services.NewAccessScope(userID, role)
	if err := h.Service.CreateOrUpdatePost(scope, post); err != nil {
		log.Printf("create post error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to save post")
	}
	normalizePostCollections(post)
	return c.Status(fiber.StatusCreated).JSON(post)
}

func (h *PostsHandler) Update(c *fiber.Ctx) error {
	userID, role, err := middleware.GetSessionUser(c, h.Store)
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
	if strings.TrimSpace(req.Text) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "text is required")
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid date")
	}
	post := &models.Post{
		ID:     int64(id),
		UserID: userID,
		Date:   date,
		Text:   req.Text,
	}
	scope := services.NewAccessScope(userID, role)
	if err := h.Service.CreateOrUpdatePost(scope, post); err != nil {
		log.Printf("update post error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update post")
	}
	normalizePostCollections(post)
	return c.JSON(post)
}

func (h *PostsHandler) Get(c *fiber.Ctx) error {
	userID, role, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	scope := services.NewAccessScope(userID, role)
	post, err := h.Service.GetPost(scope, int64(id))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "not found")
	}
	normalizePostCollections(post)
	return c.JSON(post)
}

func (h *PostsHandler) Delete(c *fiber.Ctx) error {
	userID, role, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	scope := services.NewAccessScope(userID, role)
	post, err := h.Service.GetPost(scope, int64(id))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "not found")
	}
	if err := h.deleteAttachmentFiles(post.Attachments); err != nil {
		log.Printf("delete post attachments error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete attachments")
	}
	if err := h.Service.DeletePost(scope, int64(id)); err != nil {
		log.Printf("delete post error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete post")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *PostsHandler) deleteAttachmentFiles(attachments []models.Attachment) error {
	var errs []error
	for _, attachment := range attachments {
		if attachment.FileName == "" {
			continue
		}
		path := filepath.Join(h.UploadDir, filepath.Base(attachment.FileName))
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			errs = append(errs, fmt.Errorf("%s: %w", attachment.FileName, err))
		}
	}
	return errors.Join(errs...)
}

func (h *PostsHandler) AddComment(c *fiber.Ctx) error {
	userID, role, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	postID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	scope := services.NewAccessScope(userID, role)
	if _, err := h.Service.GetPost(scope, int64(postID)); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "post not found")
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	if strings.TrimSpace(req.Text) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "text is required")
	}
	comment := &models.Comment{PostID: int64(postID), UserID: userID, Text: req.Text}
	if err := h.Service.AddComment(comment); err != nil {
		log.Printf("add comment error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to add comment")
	}
	return c.Status(fiber.StatusCreated).JSON(comment)
}

func (h *PostsHandler) UpdateComment(c *fiber.Ctx) error {
	userID, role, err := middleware.GetSessionUser(c, h.Store)
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
	if strings.TrimSpace(req.Text) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "text is required")
	}
	comment := &models.Comment{ID: int64(commentID), UserID: userID, Text: req.Text}
	scope := services.NewAccessScope(userID, role)
	if err := h.Service.UpdateComment(scope, comment); err != nil {
		log.Printf("update comment error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update comment")
	}
	return c.JSON(comment)
}

func (h *PostsHandler) DeleteComment(c *fiber.Ctx) error {
	userID, role, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	commentID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	scope := services.NewAccessScope(userID, role)
	if err := h.Service.DeleteComment(scope, int64(commentID)); err != nil {
		log.Printf("delete comment error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete comment")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *PostsHandler) ListHashtags(c *fiber.Ctx) error {
	userID, role, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	scope := services.NewAccessScope(userID, role)
	tags, err := h.Service.ListHashtags(scope)
	if err != nil {
		log.Printf("list hashtags error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list hashtags")
	}
	return c.JSON(tags)
}

func (h *PostsHandler) UploadAttachment(c *fiber.Ctx) error {
	userID, role, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	postID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	scope := services.NewAccessScope(userID, role)
	if _, err := h.Service.GetPost(scope, int64(postID)); err != nil {
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
		if !isAllowedType(contentType, h.AllowedTypes) {
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
		}
		if err := h.Service.CreateAttachment(&attachment); err != nil {
			if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
				log.Printf("rollback attachment file error: %v", removeErr)
			}
			log.Printf("create attachment error: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "failed to save attachment")
		}
		saved = append(saved, attachment)
	}
	return c.Status(fiber.StatusCreated).JSON(saved)
}

func (h *PostsHandler) DownloadAttachmentByID(c *fiber.Ctx) error {
	userID, role, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	scope := services.NewAccessScope(userID, role)
	attachment, err := h.Service.GetAttachmentByIDForUser(scope, int64(id))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "not found")
	}
	path, err := secureUploadPath(h.UploadDir, attachment.FileName)
	if err != nil {
		return fiber.NewError(fiber.StatusForbidden, "forbidden")
	}
	c.Set(fiber.HeaderXContentTypeOptions, "nosniff")
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%q", filepath.Base(attachment.FileName)))
	c.Type(attachment.FileType)
	return c.SendFile(path)
}

func isAllowedType(contentType string, allowed []string) bool {
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
	ext, err := extensionForType(contentType)
	if err != nil {
		log.Printf("warning: could not determine extension for content type %s: %v", contentType, err)
		ext = ""
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
		return "", errors.New("unsupported content type")
	}
	return extensions[0], nil
}

func secureUploadPath(uploadDir, fileName string) (string, error) {
	cleanDir := filepath.Clean(uploadDir)
	cleanFileName := filepath.Clean(fileName)
	if cleanFileName == "." || filepath.IsAbs(cleanFileName) {
		return "", errors.New("invalid file path")
	}
	fullPath := filepath.Join(cleanDir, cleanFileName)
	rel, err := filepath.Rel(cleanDir, fullPath)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, "..") || rel == "." || filepath.IsAbs(rel) {
		return "", errors.New("path traversal detected")
	}
	return fullPath, nil
}
