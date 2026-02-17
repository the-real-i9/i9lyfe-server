package helpers

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/vmihailenco/msgpack/v5"
)

func StructToMap(val any) (dest map[string]any) {
	bt, err := msgpack.Marshal(val)
	if err != nil {
		LogError(err)
	}

	if err := msgpack.Unmarshal(bt, &dest); err != nil {
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
	c.Value = base64.RawURLEncoding.EncodeToString(ToBtMsgPack(kvPairs))
	c.Path = path
	c.MaxAge = maxAge

	return c
}

func ToMsgPack(data any) string {
	d, err := msgpack.Marshal(data)
	if err != nil {
		LogError(err)
	}

	return string(d)
}

func ToBtMsgPack(data any) []byte {
	d, err := msgpack.Marshal(data)
	if err != nil {
		LogError(err)
	}

	return d
}

func ToJson(data any) string {
	d, err := json.Marshal(data)
	if err != nil {
		LogError(err)
	}

	return string(d)
}

func FromMsgPack[T any](msgPackStr string) (res T) {
	if msgPackStr == "" {
		return res
	}

	err := msgpack.Unmarshal([]byte(msgPackStr), &res)
	if err != nil {
		LogError(err)
	}

	return
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

func FromBtMsgPack[T any](msgPackBt []byte) (res T) {
	if msgPackBt == nil {
		return res
	}

	err := msgpack.Unmarshal((msgPackBt), &res)
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

func CoalesceInt(input int64, def int64) int64 {
	if input == 0 {
		return def
	}

	return input
}
