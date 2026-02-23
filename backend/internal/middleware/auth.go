package middleware

import (
	"errors"

	"familyjournal/backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func RequireAuth(store *session.Store, svc *services.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		userID, ok := sess.Get("user_id").(int64)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		user, err := svc.GetUserByID(userID)
		if err != nil || !user.IsActive {
			_ = sess.Destroy()
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
	id, idOk := idValue.(int64)
	role, roleOk := roleValue.(string)
	if !idOk || !roleOk {
		return 0, "", errors.New("invalid user session data")
	}
	return id, role, nil
}
