package handlers

import (
	"log"

	"familyjournal/backend/internal/middleware"
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
	if err := h.establishSession(c, user.ID, user.Role); err != nil {
		return err
	}
	return c.JSON(user)
}

func (h *AuthHandler) establishSession(c *fiber.Ctx, userID int64, role string) error {
	sess, err := h.Store.Get(c)
	if err != nil {
		log.Printf("session get error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "session error")
	}
	if err := sess.Regenerate(); err != nil {
		log.Printf("session regenerate error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "session error")
	}
	sess.Set("user_id", userID)
	sess.Set("role", role)
	if err := sess.Save(); err != nil {
		log.Printf("session save error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "session error")
	}
	return nil
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	sess, err := h.Store.Get(c)
	if err != nil {
		log.Printf("session get error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "session error")
	}
	if err := sess.Destroy(); err != nil {
		log.Printf("session destroy error: %v", err)
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
	var req struct {
		Email           string `json:"email"`
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	if req.Email == "" && req.NewPassword == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email or newPassword required")
	}
	if req.Email != "" {
		if err := h.Service.UpdateUserProfile(userID, req.Email); err != nil {
			log.Printf("update profile error: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "failed to update profile")
		}
	}
	if req.NewPassword != "" {
		if req.CurrentPassword == "" {
			return fiber.NewError(fiber.StatusBadRequest, "currentPassword required")
		}
		if err := h.Service.ChangePassword(userID, req.CurrentPassword, req.NewPassword); err != nil {
			log.Printf("change password error: %v", err)
			return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
		}
	}
	user, err := h.Service.GetUserByID(userID)
	if err != nil {
		log.Printf("get profile error: %v", err)
		return fiber.NewError(fiber.StatusNotFound, "not found")
	}
	return c.JSON(user)
}
