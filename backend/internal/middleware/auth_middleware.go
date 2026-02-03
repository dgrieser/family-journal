package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/user/family-journal/internal/services"
)

func AuthRequired(store *session.Store, authService *services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		val := sess.Get("user_id")
		if val == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		userID, ok := val.(uint)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user id in session"})
		}

		user, err := authService.GetUserByID(userID)
		if err != nil || !user.IsActive {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "account deactivated or not found"})
		}

		c.Locals("user_id", userID)
		c.Locals("role", string(user.Role))
		return c.Next()
	}
}

func AdminRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Next()
	}
}
