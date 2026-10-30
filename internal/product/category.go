package product

import (
	"database/sql"
	"github.com/gofiber/fiber/v2"
	"strings"
)

type CategoryHandler struct{ DB *sql.DB }

func (h *CategoryHandler) List(c *fiber.Ctx) error {
	rows, err := h.DB.QueryContext(c.Context(), `SELECT id,name,slug FROM categories WHERE active=true ORDER BY name`)
	if err != nil {
		return fiber.NewError(500, "could not list categories")
	}
	defer rows.Close()
	items := make([]fiber.Map, 0)
	for rows.Next() {
		var id, name, slug string
		if err := rows.Scan(&id, &name, &slug); err != nil {
			return fiber.NewError(500, "could not read categories")
		}
		items = append(items, fiber.Map{"id": id, "name": name, "slug": slug})
	}
	return c.JSON(items)
}

func (h *CategoryHandler) Create(c *fiber.Ctx) error {
	var input struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(400, "invalid request body")
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return fiber.NewError(400, "category name is required")
	}
	slug := strings.ToLower(strings.ReplaceAll(input.Name, " ", "-"))
	var id string
	if err := h.DB.QueryRowContext(c.Context(), `INSERT INTO categories(name,slug) VALUES($1,$2) RETURNING id`, input.Name, slug).Scan(&id); err != nil {
		return fiber.NewError(409, "category already exists or could not be created")
	}
	return c.Status(201).JSON(fiber.Map{"id": id, "name": input.Name, "slug": slug})
}
