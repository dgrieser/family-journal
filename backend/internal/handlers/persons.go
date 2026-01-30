package handlers

import (
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
	userID, _, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	persons, err := h.Service.ListPersons(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
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
	person, err := h.Service.CreatePerson(userID, req.Name, req.Description)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(person)
}

func (h *PersonsHandler) Update(c *fiber.Ctx) error {
	userID, _, err := middleware.GetSessionUser(c, h.Store)
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
	person := &models.Person{ID: int64(id), Name: req.Name, Description: req.Description}
	if err := h.Service.UpdatePerson(userID, person); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(person)
}

func (h *PersonsHandler) Delete(c *fiber.Ctx) error {
	userID, _, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	if err := h.Service.DeletePerson(userID, int64(id)); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
