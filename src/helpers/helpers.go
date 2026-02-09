package helpers

import (
	"fmt"
	"os"

	"github.com/goccy/go-json"

	"github.com/gofiber/fiber/v2"
)

func StructToMap(val any) (dest map[string]any) {
	bt, err := json.Marshal(val)
	if err != nil {
		LogError(err)
	}

	if err := json.Unmarshal(bt, &dest); err != nil {
		LogError(err)
	}

	return
}

func WSErrReply(err error, toAction string) map[string]any {

	errCode := fiber.StatusInternalServerError

	if ferr, ok := err.(*fiber.Error); ok {
		errCode = ferr.Code
	}

	errResp := map[string]any{
		"event":    "server error",
		"toAction": toAction,
		"data": map[string]any{
			"statusCode": errCode,
			"errorMsg":   fmt.Sprint(err),
		},
	}

	return errResp
}

func WSReply(data any, toAction string) map[string]any {

	reply := map[string]any{
		"event":    "server reply",
		"toAction": toAction,
		"data":     data,
	}

	return reply
}

func Session(kvPairs map[string]any, path string, maxAge int) *fiber.Cookie {
	c := &fiber.Cookie{
		HTTPOnly: true,
		Secure:   os.Getenv("GO_ENV") == "production",
		Domain:   os.Getenv("SERVER_HOST"),
	}

	c.Name = "session"
	c.Value = ToJson(kvPairs)
	c.Path = path
	c.MaxAge = maxAge

	return c
}

func ToJson(data any) string {
	d, err := json.Marshal(data)
	if err != nil {
		LogError(err)
	}

	return string(d)
}

func FromJson[T any](jsonStr string) (res T) {
	if jsonStr == "" {
		return res
	}

	err := json.Unmarshal([]byte(jsonStr), &res)
	if err != nil {
		LogError(err)
	}

	return
}

func FromBtJson[T any](jsonBt []byte) (res T) {
	if jsonBt == nil {
		return res
	}

	err := json.Unmarshal((jsonBt), &res)
	if err != nil {
		LogError(err)
	}

	return
}

func BuildNotification(notifId, notifType string, at int64, details map[string]any) map[string]any {
	return map[string]any{
		"id":      notifId,
		"type":    notifType,
		"at":      at,
		"details": details,
	}
}

func MaxCursor(cursor float64) string {
	if cursor == 0 {
		return "+inf"
	}

	return fmt.Sprintf("(%f", cursor)
}
