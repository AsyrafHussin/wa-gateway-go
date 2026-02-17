package handler

import (
	"bytes"
	"encoding/csv"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/AsyrafHussin/wa-gateway-go/internal/whatsapp"
	"github.com/AsyrafHussin/wa-gateway-go/pkg/response"
)

type Contact struct {
	manager *whatsapp.DeviceManager
	logger  zerolog.Logger
}

func NewContact(manager *whatsapp.DeviceManager, logger zerolog.Logger) *Contact {
	return &Contact{manager: manager, logger: logger}
}

func (h *Contact) List(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return response.Error(c, fiber.StatusBadRequest, "MISSING_TOKEN", "Token is required")
	}

	session, ok := h.manager.GetSession(token)
	if !ok {
		return response.Error(c, fiber.StatusNotFound, "DEVICE_NOT_FOUND", "Device not found")
	}

	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	if limit < 1 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	if offset < 0 {
		offset = 0
	}

	contacts, total, err := session.Contacts.GetAll(limit, offset)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", "Failed to retrieve contacts")
	}

	// CSV export
	if c.Query("format") == "csv" {
		var buf bytes.Buffer
		w := csv.NewWriter(&buf)

		_ = w.Write([]string{"phone", "name", "source", "first_seen", "last_seen"})
		for _, contact := range contacts {
			_ = w.Write([]string{
				contact.Phone,
				contact.Name,
				contact.Source,
				contact.FirstSeen,
				contact.LastSeen,
			})
		}
		w.Flush()

		if err := w.Error(); err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "CSV_ERROR", "Failed to generate CSV")
		}

		c.Set("Content-Type", "text/csv")
		c.Set("Content-Disposition", "attachment; filename=contacts-"+token+".csv")
		return c.Send(buf.Bytes())
	}

	return response.Success(c, fiber.StatusOK, fiber.Map{
		"contacts": contacts,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	}, "Contacts retrieved")
}
