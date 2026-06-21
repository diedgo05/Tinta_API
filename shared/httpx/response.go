// Package httpx provides standardized HTTP response helpers for all Tinta services.
package httpx

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

// ErrorBody is the standard error payload returned by every service.
type ErrorBody struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details any    `json:"details,omitempty"`
}

// OK writes a 200 response with the given data.
func OK(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(data)
}

// Created writes a 201 response with the given data.
func Created(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(data)
}

// NoContent writes a 204 response without body.
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// Error writes a JSON error response with the given status and message.
func Error(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(ErrorBody{
		Error: message,
		Code:  code,
	})
}

// ErrorFromFiber unwraps a fiber.Error and writes its JSON payload.
// If err is not a fiber.Error, returns a 500 with the error message.
func ErrorFromFiber(c *fiber.Ctx, err error) error {
	var fe *fiber.Error
	if errors.As(err, &fe) {
		return c.Status(fe.Code).JSON(ErrorBody{Error: fe.Message})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(ErrorBody{Error: err.Error()})
}
