package handlers

import (
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
	persons, err := h.repo.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(persons)
}

func (h *PersonHandler) Create(c *fiber.Ctx) error {
	var person models.Person
	if err := c.BodyParser(&person); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse json"})
	}

	userID := c.Locals("user_id").(uint)
	person.CreatedByUserID = userID

	if err := h.repo.Create(&person); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(person)
}

func (h *PersonHandler) Update(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	var person models.Person
	if err := c.BodyParser(&person); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse json"})
	}

	person.ID = uint(id)
	if err := h.repo.Update(&person); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(person)
}

func (h *PersonHandler) Delete(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	if err := h.repo.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
