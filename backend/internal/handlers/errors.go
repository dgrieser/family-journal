package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func JSONErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "internal server error"

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		code = fiberErr.Code
		if fiberErr.Message != "" {
			message = fiberErr.Message
		}
	} else if err != nil {
		message = err.Error()
	}

	message = strings.TrimSpace(message)
	if message == "" {
		message = "internal server error"
	}

	return c.Status(code).JSON(fiber.Map{"error": message})
}
