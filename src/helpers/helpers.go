package helpers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"

	"github.com/gofiber/fiber/v2"
)

func LogError(err error) {
	if err == nil {
		return
	}

	pc, file, line, ok := runtime.Caller(1)
	fn := "unknown"
	if !ok {
		file = "???"
		line = 0
	} else {
		fn = runtime.FuncForPC(pc).Name()
	}

	log.Printf("[ERROR] %s:%d %s(): %v\n", file, line, fn, err)
}

func ToStruct(val any, dest any) {
	valType := reflect.TypeOf(val)

	if valType.Kind() != reflect.Map && !(valType.Kind() == reflect.Slice && valType.Elem().Kind() == reflect.Map) {
		panic("expected 'val' to be a map or slice of maps")
	}

	destType := reflect.TypeOf(dest).Elem()

	if destType.Kind() != reflect.Struct && !(destType.Kind() == reflect.Slice && destType.Elem().Kind() == reflect.Struct) {
		panic("expected 'dest' to be a pointer to struct or slice of structs")
	}

	bt, err := json.Marshal(val)
	if err != nil {
		LogError(err)
	}

	if err := json.Unmarshal(bt, dest); err != nil {
		LogError(err)
	}
}

func StructToMap(val any) (dest map[string]any) {
	if reflect.TypeOf(val).Kind() != reflect.Struct {
		panic("expected 'val' to be a struct")
	}

	var mp map[string]any

	bt, err := json.Marshal(val)
	if err != nil {
		LogError(err)
	}

	if err := json.Unmarshal(bt, &mp); err != nil {
		LogError(err)
	}

	return mp
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
		Secure:   false,
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
