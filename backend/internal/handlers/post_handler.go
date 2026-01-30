package handlers

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/user/family-journal/internal/services"
)

type PostHandler struct {
	postService *services.PostService
}

func NewPostHandler(postService *services.PostService) *PostHandler {
	return &PostHandler{postService: postService}
}

func (h *PostHandler) GetPosts(c *fiber.Ctx) error {
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

	posts, err := h.postService.GetPosts(date, hTags, pNames, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(posts)
}

func (h *PostHandler) GetPost(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	post, err := h.postService.GetPost(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "post not found"})
	}
	return c.JSON(post)
}

type CreatePostRequest struct {
	Date string `json:"date"`
	Text string `json:"text"`
}

func (h *PostHandler) Create(c *fiber.Ctx) error {
	var req CreatePostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse json"})
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		date = time.Now()
	}

	userID := c.Locals("user_id").(uint)
	post, err := h.postService.CreatePost(userID, date, req.Text)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Handle file uploads
	form, err := c.MultipartForm()
	if err == nil {
		files := form.File["attachments"]
		for _, file := range files {
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

			ext := filepath.Ext(file.Filename)
			newFileName := uuid.New().String() + ext
			storagePath := filepath.Join("uploads", newFileName)

			if err := c.SaveFile(file, storagePath); err != nil {
				continue
			}

			h.postService.AddAttachment(post.ID, file.Filename, file.Header.Get("Content-Type"), file.Size, storagePath)
		}
	}

	return c.Status(fiber.StatusCreated).JSON(post)
}

func (h *PostHandler) Update(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	var req CreatePostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse json"})
	}

	post, err := h.postService.UpdatePost(uint(id), req.Text)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(post)
}

func (h *PostHandler) Delete(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	if err := h.postService.DeletePost(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type CommentRequest struct {
	Text string `json:"text"`
}

func (h *PostHandler) AddComment(c *fiber.Ctx) error {
	postID, _ := c.ParamsInt("id")
	var req CommentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse json"})
	}

	userID := c.Locals("user_id").(uint)
	comment, err := h.postService.AddComment(uint(postID), userID, req.Text)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(comment)
}

func (h *PostHandler) DeleteComment(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id") // this would be comment id
	if err := h.postService.DeleteComment(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *PostHandler) DownloadAttachment(c *fiber.Ctx) error {
	path := c.Query("path")
	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "path is required"})
	}
	return c.Download(path)
}

func (h *PostHandler) GetHashtags(c *fiber.Ctx) error {
	hashtags, err := h.postService.GetAllHashtags()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(hashtags)
}
