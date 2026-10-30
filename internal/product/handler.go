package product

import (
	"database/sql"
	"github.com/gofiber/fiber/v2"
	"strconv"
	"strings"
)

type Handler struct{ DB *sql.DB }

type productInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CategoryID  string `json:"category_id"`
	Price       int64  `json:"price"`
	Stock       int    `json:"stock"`
}

func (h *Handler) List(c *fiber.Ctx) error {
	rows, err := h.DB.QueryContext(c.Context(), `SELECT id,name,description,category_id,price,stock FROM products WHERE active=true ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		return fiber.NewError(500, "could not list products")
	}
	defer rows.Close()
	result := make([]fiber.Map, 0)
	for rows.Next() {
		var id, name, description, category string
		var price int64
		var stock int
		if err := rows.Scan(&id, &name, &description, &category, &price, &stock); err != nil {
			return fiber.NewError(500, "could not read products")
		}
		result = append(result, fiber.Map{"id": id, "name": name, "description": description, "category_id": category, "price": price, "stock": stock})
	}
	if err := rows.Err(); err != nil {
		return fiber.NewError(500, "could not read products")
	}
	return c.JSON(result)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	var p productInput
	var id string
	err := h.DB.QueryRowContext(c.Context(), `SELECT id,name,description,category_id,price,stock FROM products WHERE id=$1 AND active=true`, c.Params("id")).Scan(&id, &p.Name, &p.Description, &p.CategoryID, &p.Price, &p.Stock)
	if err == sql.ErrNoRows {
		return fiber.NewError(404, "product not found")
	}
	if err != nil {
		return fiber.NewError(500, "could not load product")
	}
	return c.JSON(fiber.Map{"id": id, "name": p.Name, "description": p.Description, "category_id": p.CategoryID, "price": p.Price, "stock": p.Stock})
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var p productInput
	if err := c.BodyParser(&p); err != nil {
		return fiber.NewError(400, "invalid request body")
	}
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" || p.Price <= 0 || p.Stock < 0 {
		return fiber.NewError(400, "name, positive price and non-negative stock are required")
	}
	var id string
	err := h.DB.QueryRowContext(c.Context(), `INSERT INTO products(name,description,category_id,price,stock) VALUES($1,$2,NULLIF($3,'')::uuid,$4,$5) RETURNING id`, p.Name, p.Description, p.CategoryID, p.Price, p.Stock).Scan(&id)
	if err != nil {
		return fiber.NewError(400, "could not create product")
	}
	return c.Status(201).JSON(fiber.Map{"id": id, "message": "product created"})
}

func (h *Handler) Update(c *fiber.Ctx) error {
	var p productInput
	if err := c.BodyParser(&p); err != nil {
		return fiber.NewError(400, "invalid request body")
	}
	if p.Name == "" || p.Price <= 0 || p.Stock < 0 {
		return fiber.NewError(400, "invalid product fields")
	}
	result, err := h.DB.ExecContext(c.Context(), `UPDATE products SET name=$1,description=$2,category_id=NULLIF($3,'')::uuid,price=$4,stock=$5,updated_at=now() WHERE id=$6 AND active=true`, p.Name, p.Description, p.CategoryID, p.Price, p.Stock, c.Params("id"))
	if err != nil {
		return fiber.NewError(400, "could not update product")
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return fiber.NewError(404, "product not found")
	}
	return c.SendStatus(204)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	result, err := h.DB.ExecContext(c.Context(), `UPDATE products SET active=false,updated_at=now() WHERE id=$1 AND active=true`, c.Params("id"))
	if err != nil {
		return fiber.NewError(500, "could not delete product")
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return fiber.NewError(404, "product not found")
	}
	return c.SendStatus(204)
}

func (h *Handler) Reserve(c *fiber.Ctx) error {
	var input struct {
		Quantity int `json:"quantity"`
	}
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(400, "invalid request body")
	}
	if input.Quantity < 1 || input.Quantity > 1000 {
		return fiber.NewError(400, "quantity must be between 1 and 1000")
	}
	var price int64
	err := h.DB.QueryRowContext(c.Context(), `UPDATE products SET stock=stock-$1,updated_at=now() WHERE id=$2 AND active=true AND stock >= $1 RETURNING price`, input.Quantity, c.Params("id")).Scan(&price)
	if err == sql.ErrNoRows {
		return fiber.NewError(409, "product unavailable or insufficient stock")
	}
	if err != nil {
		return fiber.NewError(500, "inventory reservation failed")
	}
	return c.JSON(fiber.Map{"reserved": true, "quantity": input.Quantity, "unit_price": price})
}

func (h *Handler) Release(c *fiber.Ctx) error {
	quantity, err := strconv.Atoi(c.Query("quantity"))
	if err != nil || quantity < 1 || quantity > 1000 {
		return fiber.NewError(400, "valid quantity is required")
	}
	result, err := h.DB.ExecContext(c.Context(), `UPDATE products SET stock=stock+$1,updated_at=now() WHERE id=$2 AND active=true`, quantity, c.Params("id"))
	if err != nil {
		return fiber.NewError(500, "inventory release failed")
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return fiber.NewError(404, "product not found")
	}
	return c.SendStatus(204)
}
