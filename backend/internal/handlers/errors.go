package handlers

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func JSONErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	var message string

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		code = fiberErr.Code
		message = fiberErr.Message
	} else if err != nil {
		log.Printf("unhandled error: %v", err)
	}

	message = strings.TrimSpace(message)
	if message == "" {
		if code >= fiber.StatusInternalServerError {
			message = "internal server error"
		} else {
			message = strings.TrimSpace(http.StatusText(code))
		}
		if message == "" {
			message = "request failed"
		}
	}

	return c.Status(code).JSON(fiber.Map{"error": message})
}
