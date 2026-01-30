package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func RequireAuth(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		userID := sess.Get("user_id")
		if userID == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		return c.Next()
	}
}

func RequireRole(store *session.Store, role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		current := sess.Get("role")
		if current != role {
			return fiber.NewError(fiber.StatusForbidden, "forbidden")
		}
		return c.Next()
	}
}

func GetSessionUser(c *fiber.Ctx, store *session.Store) (int64, string, error) {
	sess, err := store.Get(c)
	if err != nil {
		return 0, "", err
	}
	idValue := sess.Get("user_id")
	roleValue := sess.Get("role")
	id, _ := idValue.(int64)
	role, _ := roleValue.(string)
	return id, role, nil
}
