package helpers

import (
	"encoding/json"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func ToStruct(val any, dest any) {
	if reflect.TypeOf(dest).Elem().Kind() != reflect.Struct {
		panic("expected 'dest' to be a struct")
	}

	if !reflect.TypeOf(val).ConvertibleTo(reflect.TypeOf(dest).Elem()) {
		panic("'val' not convertible to 'dest'")
	}

	bt, err := json.Marshal(val)
	if err != nil {
		log.Println("helpers.go: ToStruct: json.Marshal:", err)
	}

	if err := json.Unmarshal(bt, dest); err != nil {
		log.Println("helpers.go: ToStruct: json.Unmarshal:", err)
	}
}

func StructToMap(val any, dest *map[string]any) {
	if reflect.TypeOf(val).Kind() != reflect.Struct {
		panic("expected 'val' to be a struct")
	}

	valNumField := reflect.TypeOf(val).NumField()

	var resMap = make(map[string]any, valNumField)

	for i := range valNumField {
		key := ""
		jsonTag := reflect.TypeOf(val).Field(i).Tag.Get("json")

		if jsonTag != "" {
			key = jsonTag
		} else {
			key = strings.ToLower(reflect.TypeOf(val).Field(i).Name)
		}

		fieldVal := reflect.ValueOf(val).Field(i).Interface()

		resMap[key] = fieldVal
	}

	*dest = resMap
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
