package helpers

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

func MapToStruct(val map[string]any, yourStruct any) {
	bt, _ := json.Marshal(val)

	if err := json.Unmarshal(bt, yourStruct); err != nil {
		log.Println("helpers.go: MapToStruct:", err)
	}
}

func AnyToAny(val any, dest any) {
	bt, _ := json.Marshal(val)

	if err := json.Unmarshal(bt, dest); err != nil {
		log.Println("helpers.go: AnyToAny:", err)
	}
}

// Includes a business-specific functionality for a default offset time
//
// When the user provides no explicit offset msec value or specifies zero,
// this implies that she wants results from the most-recent content
// (note that results are returned in descending order).
// However, converting a zero msec into time.Time yields a past time,
// and since we'll normally fetch contents whose time created is less than
// the offset specified, we therefore need to coalesce the past time into future time.
func OffsetTime(msec int64) time.Time {
	if msec == 0 {
		return time.Now().Add(time.Minute).UTC()
	}

	return time.UnixMilli(msec).UTC()
}

func WSErrReply(err error, onEvent string) map[string]any {

	errCode := fiber.StatusInternalServerError

	if ferr, ok := err.(*fiber.Error); ok {
		errCode = ferr.Code
	}

	errResp := map[string]any{
		"event":   "server error",
		"onEvent": onEvent,
		"data": map[string]any{
			"statusCode": errCode,
			"errorMsg":   err.Error(),
		},
	}

	return errResp
}

func WSReply(data any, onEvent string) map[string]any {

	reply := map[string]any{
		"event":   "server reply",
		"onEvent": onEvent,
		"data":    data,
	}

	return reply
}

func Cookie(name, value, path string, maxAge int) *fiber.Cookie {
	c := &fiber.Cookie{
		HTTPOnly: true,
		Secure:   false,
		Domain:   os.Getenv("SERVER_HOST"),
	}

	c.Name = name
	c.Value = value
	c.Path = path
	c.MaxAge = maxAge

	return c
}
