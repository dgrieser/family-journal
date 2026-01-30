package handlers

import (
	"familyjournal/backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
	Service *services.Service
}

type roleRequest struct {
	Role string `json:"role"`
}

type activeRequest struct {
	Active bool `json:"active"`
}

func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	users, err := h.Service.ListUsers()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(users)
}

func (h *AdminHandler) UpdateRole(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req roleRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	if req.Role != "admin" && req.Role != "user" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid role")
	}
	if err := h.Service.UpdateUserRole(int64(id), req.Role); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AdminHandler) UpdateActive(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req activeRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	if err := h.Service.UpdateUserActive(int64(id), req.Active); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
