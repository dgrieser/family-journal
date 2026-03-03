package handlers

import (
	"familyjournal/backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

func parsePagination(c *fiber.Ctx) (services.PaginationParams, error) {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", services.DefaultPageSize)
	return services.NewPagination(page, pageSize), nil
}
