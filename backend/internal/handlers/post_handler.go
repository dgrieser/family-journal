package handlers

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/user/family-journal/internal/models"
	"github.com/user/family-journal/internal/services"
)

type PostHandler struct {
	postService *services.PostService
}

func NewPostHandler(postService *services.PostService) *PostHandler {
	return &PostHandler{postService: postService}
}

func (h *PostHandler) GetPosts(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user id in session"})
	}

	dateStr := c.Query("date")
	hashtags := c.Query("hashtags")
	persons := c.Query("persons")
	search := c.Query("search")

	var date *time.Time
	if dateStr != "" {
		d, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			date = &d
		}
	}

	var hTags []string
	if hashtags != "" {
		hTags = strings.Split(hashtags, ",")
	}

	var pNames []string
	if persons != "" {
		pNames = strings.Split(persons, ",")
	}

	posts, err := h.postService.GetPosts(userID, date, hTags, pNames, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(posts)
}

func (h *PostHandler) GetPost(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id parameter"})
	}

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user id in session"})
	}

	post, err := h.postService.GetPost(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "post not found"})
	}

	if post.UserID != userID && !isAdmin(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "unauthorized"})
	}

	return c.JSON(post)
}

type CreatePostRequest struct {
	Date string `json:"date"`
	Text string `json:"text"`
}

type AttachmentResponse struct {
	ID        uint      `json:"id"`
	PostID    uint      `json:"post_id"`
	FileName  string    `json:"file_name"`
	FileType  string    `json:"file_type"`
	FileSize  int64     `json:"file_size"`
	CreatedAt time.Time `json:"created_at"`
}

func toAttachmentResponse(a *models.Attachment) AttachmentResponse {
	return AttachmentResponse{
		ID:        a.ID,
		PostID:    a.PostID,
		FileName:  a.FileName,
		FileType:  a.FileType,
		FileSize:  a.FileSize,
		CreatedAt: a.CreatedAt,
	}
}

func isAdmin(c *fiber.Ctx) bool {
	role, ok := c.Locals("role").(string)
	return ok && role == "admin"
}

func (h *PostHandler) handleFileUploads(c *fiber.Ctx, postID uint) ([]AttachmentResponse, error) {
	if !strings.HasPrefix(c.Get(fiber.HeaderContentType), fiber.MIMEMultipartForm) {
		return nil, nil
	}

	form, err := c.MultipartForm()
	if err != nil {
		return nil, fmt.Errorf("invalid multipart form: %w", err)
	}

	files := form.File["files"]
	if len(files) == 0 {
		// Backward compatibility for older clients.
		files = form.File["attachments"]
	}

	uploaded := make([]AttachmentResponse, 0, len(files))
	for _, file := range files {
		// Validate file extension
		ext := strings.ToLower(filepath.Ext(file.Filename))
		allowedExts := map[string]bool{
			".jpg":  true,
			".jpeg": true,
			".png":  true,
			".pdf":  true,
		}
		if !allowedExts[ext] {
			continue
		}

		// Validate file type
		contentType := file.Header.Get("Content-Type")
		allowedTypes := map[string]bool{
			"image/jpeg":      true,
			"image/png":       true,
			"application/pdf": true,
		}
		if !allowedTypes[contentType] {
			continue
		}

		// Validate file size (e.g. 5MB)
		if file.Size > 5*1024*1024 {
			continue
		}

		newFileName := uuid.New().String() + ext
		storagePath := filepath.Join("uploads", newFileName)

		if err := c.SaveFile(file, storagePath); err != nil {
			return nil, fmt.Errorf("save file %q: %w", file.Filename, err)
		}

		attachment, err := h.postService.AddAttachment(postID, file.Filename, contentType, file.Size, storagePath)
		if err != nil {
			return nil, fmt.Errorf("persist attachment %q: %w", file.Filename, err)
		}
		uploaded = append(uploaded, toAttachmentResponse(attachment))
	}
	return uploaded, nil
}

func (h *PostHandler) AddAttachments(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id parameter"})
	}

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user id in session"})
	}

	post, err := h.postService.GetPost(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "post not found"})
	}

	if post.UserID != userID && !isAdmin(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "unauthorized"})
	}

	attachments, err := h.handleFileUploads(c, post.ID)
	if err != nil {
		log.Printf("failed to upload attachments for post %d: %v", post.ID, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to process file upload"})
	}

	return c.JSON(attachments)
}

func (h *PostHandler) Create(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user id in session"})
	}

	req := CreatePostRequest{
		Text: c.FormValue("text"),
		Date: c.FormValue("date"),
	}

	if req.Text == "" || req.Date == "" {
		// Fallback to JSON if form values are empty (for backward compatibility or direct API usage)
		_ = c.BodyParser(&req)
	}

	if req.Text == "" || req.Date == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "text and date are required"})
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid date format, please use YYYY-MM-DD"})
	}

	post, err := h.postService.CreatePost(userID, date, req.Text)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if _, err := h.handleFileUploads(c, post.ID); err != nil {
		log.Printf("failed to upload attachments for post %d: %v", post.ID, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to process file upload"})
	}

	return c.Status(fiber.StatusCreated).JSON(post)
}

func (h *PostHandler) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id parameter"})
	}

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user id in session"})
	}

	existingPost, err := h.postService.GetPost(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "post not found"})
	}

	if existingPost.UserID != userID && !isAdmin(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "unauthorized"})
	}

	req := CreatePostRequest{
		Text: c.FormValue("text"),
		Date: c.FormValue("date"),
	}

	if req.Text == "" && req.Date == "" {
		_ = c.BodyParser(&req)
	}

	var postDate *time.Time
	if req.Date != "" {
		d, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid date format, please use YYYY-MM-DD"})
		}
		postDate = &d
	}

	post, err := h.postService.UpdatePost(uint(id), req.Text, postDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if _, err := h.handleFileUploads(c, post.ID); err != nil {
		log.Printf("failed to upload attachments for post %d: %v", post.ID, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to process file upload"})
	}

	return c.JSON(post)
}

func (h *PostHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id parameter"})
	}

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user id in session"})
	}

	existingPost, err := h.postService.GetPost(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "post not found"})
	}

	if existingPost.UserID != userID && !isAdmin(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "unauthorized"})
	}

	if err := h.postService.DeletePost(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type CommentRequest struct {
	Text string `json:"text"`
}

func (h *PostHandler) AddComment(c *fiber.Ctx) error {
	postID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id parameter"})
	}
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user id in session"})
	}

	post, err := h.postService.GetPost(uint(postID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "post not found"})
	}

	// Verify post ownership or admin role
	if post.UserID != userID && !isAdmin(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req CommentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse json"})
	}

	comment, err := h.postService.AddComment(uint(postID), userID, req.Text)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(comment)
}

func (h *PostHandler) DeleteComment(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id parameter"})
	}

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user id in session"})
	}

	existingComment, err := h.postService.GetComment(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "comment not found"})
	}

	if existingComment.UserID != userID && !isAdmin(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "unauthorized"})
	}

	if err := h.postService.DeleteComment(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *PostHandler) DownloadAttachment(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id is required"})
	}

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user id in session"})
	}

	attachment, err := h.postService.GetAttachment(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "attachment not found"})
	}

	post, err := h.postService.GetPost(attachment.PostID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "associated post not found"})
	}

	if post.UserID != userID && !isAdmin(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "unauthorized"})
	}

	// Extra safety: check if file is within uploads directory
	cleanPath := filepath.Clean(attachment.StoragePath)
	if !strings.HasPrefix(cleanPath, "uploads") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}
	return c.SendFile(cleanPath)
}

func (h *PostHandler) GetHashtags(c *fiber.Ctx) error {
	hashtags, err := h.postService.GetAllHashtags()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(hashtags)
}
