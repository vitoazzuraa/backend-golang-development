package helper

import (
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"backend-go/api-student/app/model"

	"github.com/gofiber/fiber/v2"
)

var ErrInvalidCursor = errors.New("cursor tidak sah")

const (
	defaultListLimit = 10
	maxListLimit     = 100
)

func EncodeCursor(createdAt time.Time, id int) string {
	raw := strconv.FormatInt(createdAt.UTC().UnixNano(), 10) +
		"|" + strconv.Itoa(id)

	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(encoded string) (model.Cursor, error) {
	if strings.TrimSpace(encoded) == "" {
		return model.Cursor{}, ErrInvalidCursor
	}

	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return model.Cursor{}, ErrInvalidCursor
	}

	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return model.Cursor{}, ErrInvalidCursor
	}

	nanos, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return model.Cursor{}, ErrInvalidCursor
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil || id < 1 {
		return model.Cursor{}, ErrInvalidCursor
	}

	return model.Cursor{CreatedAt: time.Unix(0, nanos).UTC(), ID: id}, nil
}

func ParseCursorQuery(c *fiber.Ctx) (model.CursorQuery, error) {
	q := model.CursorQuery{
		Limit:  c.QueryInt("limit", defaultListLimit),
		Search: strings.TrimSpace(c.Query("search")),
	}

	if q.Limit < 1 {
		q.Limit = defaultListLimit
	}
	if q.Limit > maxListLimit {
		q.Limit = maxListLimit
	}

	if raw := c.Query("is_active"); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return model.CursorQuery{}, BadRequest("is_active harus bernilai true atau false")
		}

		q.IsActive = &value
	}

	if raw := strings.TrimSpace(c.Query("cursor")); raw != "" {
		cursor, err := DecodeCursor(raw)
		if err != nil {
			return model.CursorQuery{}, BadRequest("cursor tidak sah")
		}

		q.After = &cursor
	}

	return q, nil
}
