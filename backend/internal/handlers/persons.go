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

type PersonsHandler struct {
	Service *services.Service
	Store   *session.Store
}

func (h *PersonsHandler) List(c *fiber.Ctx) error {
	userID, role, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	search := strings.TrimSpace(c.Query("search"))
	pagination := parsePagination(c)
	scope := services.NewAccessScope(userID, role)
	persons, err := h.Service.ListPersons(scope, search, pagination)
	if err != nil {
		log.Printf("list persons error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list persons")
	}
	return c.JSON(persons)
}

func (h *PersonsHandler) Create(c *fiber.Ctx) error {
	userID, _, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	if strings.TrimSpace(req.Name) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	person, err := h.Service.CreatePerson(userID, req.Name, req.Description)
	if err != nil {
		if errors.Is(err, models.ErrDuplicate) {
			return fiber.NewError(fiber.StatusConflict, "person already exists")
		}
		log.Printf("create person error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create person")
	}
	return c.Status(fiber.StatusCreated).JSON(person)
}

func (h *PersonsHandler) Update(c *fiber.Ctx) error {
	userID, role, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	if strings.TrimSpace(req.Name) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	person := &models.Person{ID: int64(id), Name: req.Name, Description: req.Description}
	scope := services.NewAccessScope(userID, role)
	if err := h.Service.UpdatePerson(scope, person); err != nil {
		if errors.Is(err, models.ErrDuplicate) {
			return fiber.NewError(fiber.StatusConflict, "person already exists")
		}
		if errors.Is(err, models.ErrForbidden) {
			return fiber.NewError(fiber.StatusForbidden, "forbidden")
		}
		log.Printf("update person error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update person")
	}
	return c.JSON(person)
}

func (h *PersonsHandler) Delete(c *fiber.Ctx) error {
	userID, role, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	scope := services.NewAccessScope(userID, role)
	if err := h.Service.DeletePerson(scope, int64(id)); err != nil {
		if errors.Is(err, models.ErrForbidden) {
			return fiber.NewError(fiber.StatusForbidden, "forbidden")
		}
		log.Printf("delete person error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete person")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
