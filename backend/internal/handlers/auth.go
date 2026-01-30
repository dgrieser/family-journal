package handlers

import (
	"familyjournal/backend/internal/middleware"
	"familyjournal/backend/internal/models"
	"familyjournal/backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type AuthHandler struct {
	Service *services.Service
	Store   *session.Store
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req authRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	user, err := h.Service.Register(req.Email, req.Password)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(user)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req authRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	user, err := h.Service.Authenticate(req.Email, req.Password)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}
	sess, err := h.Store.Get(c)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "session error")
	}
	sess.Set("user_id", user.ID)
	sess.Set("role", user.Role)
	if err := sess.Save(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "session error")
	}
	return c.JSON(user)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	sess, err := h.Store.Get(c)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "session error")
	}
	if err := sess.Destroy(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "session error")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) Profile(c *fiber.Ctx) error {
	userID, _, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	user, err := h.Service.GetUserByID(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "not found")
	}
	return c.JSON(user)
}

func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, _, err := middleware.GetSessionUser(c, h.Store)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	var req models.User
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	if req.Email == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email required")
	}
	if err := h.Service.UpdateUserProfile(userID, req.Email); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	user, err := h.Service.GetUserByID(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "not found")
	}
	return c.JSON(user)
}
