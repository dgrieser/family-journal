package handlers

import (
	"log"

	"familyjournal/backend/internal/models"
	"familyjournal/backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
	Service *services.Service
}

type roleRequest struct {
	Role string `json:"role"`
}

type updateActiveRequest struct {
	IsActive bool `json:"is_active"`
}

func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	users, err := h.Service.ListUsers()
	if err != nil {
		log.Printf("list users error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to retrieve users")
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
	if req.Role != models.RoleAdmin && req.Role != models.RoleUser {
		return fiber.NewError(fiber.StatusBadRequest, "invalid role")
	}
	if err := h.Service.UpdateUserRole(int64(id), req.Role); err != nil {
		log.Printf("update user role error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update role")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AdminHandler) UpdateActive(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req updateActiveRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	if err := h.Service.UpdateUserActive(int64(id), req.IsActive); err != nil {
		log.Printf("update user active error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update user")
	}
	user, err := h.Service.GetUserByID(int64(id))
	if err != nil {
		log.Printf("get user after active update error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to retrieve user")
	}
	return c.JSON(user)
}
