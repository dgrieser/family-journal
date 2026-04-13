package handlers

import (
	"errors"
	"log"
	"strings"

	"familyjournal/backend/internal/middleware"
	"familyjournal/backend/internal/models"
	"familyjournal/backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type HashtagsHandler struct {
	Service *services.Service
	Store   *session.Store
}

func (h *HashtagsHandler) List(c *fiber.Ctx) error {
	_, _, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	tags, err := h.Service.ListAllHashtags()
	if err != nil {
		log.Printf("list hashtags error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list hashtags")
	}
	return c.JSON(tags)
}

func (h *HashtagsHandler) Create(c *fiber.Ctx) error {
	userID, _, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	if strings.TrimSpace(req.Name) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	tag, err := h.Service.CreateHashtag(userID, req.Name)
	if err != nil {
		if errors.Is(err, models.ErrDuplicate) {
			return fiber.NewError(fiber.StatusConflict, "hashtag already exists")
		}
		log.Printf("create hashtag error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create hashtag")
	}
	return c.Status(fiber.StatusCreated).JSON(tag)
}

func (h *HashtagsHandler) Update(c *fiber.Ctx) error {
	userID, role, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	if strings.TrimSpace(req.Name) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	tag := &models.Hashtag{ID: int64(id), Name: req.Name}
	scope := services.NewAccessScope(userID, role)
	if err := h.Service.UpdateHashtag(scope, tag); err != nil {
		if errors.Is(err, models.ErrDuplicate) {
			return fiber.NewError(fiber.StatusConflict, "hashtag already exists")
		}
		if errors.Is(err, models.ErrForbidden) {
			return fiber.NewError(fiber.StatusForbidden, "forbidden")
		}
		log.Printf("update hashtag error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update hashtag")
	}
	return c.JSON(tag)
}

func (h *HashtagsHandler) Delete(c *fiber.Ctx) error {
	userID, role, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	scope := services.NewAccessScope(userID, role)
	if err := h.Service.DeleteHashtag(scope, int64(id)); err != nil {
		if errors.Is(err, models.ErrForbidden) {
			return fiber.NewError(fiber.StatusForbidden, "forbidden")
		}
		log.Printf("delete hashtag error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete hashtag")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
