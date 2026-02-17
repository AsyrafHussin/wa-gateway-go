package handler

import (
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
	if limit > 1000 {
		limit = 1000
	}

	contacts, total, err := session.Contacts.GetAll(limit, offset)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error())
	}

	// CSV export
	if c.Query("format") == "csv" {
		c.Set("Content-Type", "text/csv")
		c.Set("Content-Disposition", "attachment; filename=contacts-"+token+".csv")
		csv := "phone,name,source,first_seen,last_seen\n"
		for _, contact := range contacts {
			csv += contact.Phone + "," + contact.Name + "," + contact.Source + "," + contact.FirstSeen + "," + contact.LastSeen + "\n"
		}
		return c.SendString(csv)
	}

	return response.Success(c, fiber.StatusOK, fiber.Map{
		"contacts": contacts,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	}, "Contacts retrieved")
}
