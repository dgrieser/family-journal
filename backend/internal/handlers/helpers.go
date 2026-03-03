package handlers

import (
	"familyjournal/backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

func parsePagination(c *fiber.Ctx) (services.PaginationParams, error) {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", services.DefaultPageSize)
	if page < 1 {
		return services.PaginationParams{}, fiber.NewError(fiber.StatusBadRequest, "invalid page")
	}
	if pageSize < 1 {
		return services.PaginationParams{}, fiber.NewError(fiber.StatusBadRequest, "invalid pageSize")
	}
	return services.NewPagination(page, pageSize), nil
}
