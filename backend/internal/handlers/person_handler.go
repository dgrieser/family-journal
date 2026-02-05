package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/user/family-journal/internal/models"
	"github.com/user/family-journal/internal/repository"
)

type PersonHandler struct {
	repo *repository.PersonRepository
}

func NewPersonHandler(repo *repository.PersonRepository) *PersonHandler {
	return &PersonHandler{repo: repo}
}

func (h *PersonHandler) GetAll(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user id in session"})
	}

	persons, err := h.repo.GetAll(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(persons)
}

type PersonRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *PersonHandler) Create(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user id in session"})
	}

	var req PersonRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse json"})
	}

	person := models.Person{
		Name:            strings.ToLower(req.Name),
		Description:     req.Description,
		CreatedByUserID: &userID,
	}

	if err := h.repo.Create(&person); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(person)
}

func (h *PersonHandler) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id parameter"})
	}

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user id in session"})
	}

	existing, err := h.repo.FindByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "person not found"})
	}

	if (existing.CreatedByUserID == nil || *existing.CreatedByUserID != userID) && c.Locals("role") != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req PersonRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse json"})
	}

	existing.Name = strings.ToLower(req.Name)
	existing.Description = req.Description

	if err := h.repo.Update(existing); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(existing)
}

func (h *PersonHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id parameter"})
	}

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user id in session"})
	}

	existing, err := h.repo.FindByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "person not found"})
	}

	if (existing.CreatedByUserID == nil || *existing.CreatedByUserID != userID) && c.Locals("role") != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "unauthorized"})
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
