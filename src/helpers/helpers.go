package helpers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"

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
	destElem := reflect.TypeOf(dest).Elem()

	if destElem.Kind() != reflect.Struct && !(destElem.Kind() == reflect.Slice && destElem.Elem().Kind() == reflect.Struct) {
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

func WSErrReply(err error, toEvent string) map[string]any {

	errCode := fiber.StatusInternalServerError

	if ferr, ok := err.(*fiber.Error); ok {
		errCode = ferr.Code
	}

	errResp := map[string]any{
		"event":   "server error",
		"toEvent": toEvent,
		"data": map[string]any{
			"statusCode": errCode,
			"errorMsg":   fmt.Sprint(err),
		},
	}

	return errResp
}

func WSReply(data any, toEvent string) map[string]any {

	reply := map[string]any{
		"event":   "server reply",
		"toEvent": toEvent,
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

func ToJson(data any) string {
	d, err := json.Marshal(data)
	if err != nil {
		LogError(err)
	}
	return string(d)
}

func Json2Map(jsonStr string) (res map[string]any) {
	err := json.Unmarshal([]byte(jsonStr), &res)
	if err != nil {
		LogError(err)
	}

	return
}

func BuildPostMentionNotification(notifId, postId, mentioningUser string, at time.Time) map[string]any {
	notif := ToJson(map[string]any{
		"type": "mention_in_post",
		"at":   at.UnixMilli(),
		"details": map[string]any{
			"in_post_id":      postId,
			"mentioning_user": mentioningUser,
		},
	})

	return map[string]any{
		"id":      notifId,
		"notif":   notif,
		"is_read": false,
	}
}

func BuildCommentMentionNotification(notifId, commentId, mentioningUser string, at time.Time) map[string]any {
	notif := ToJson(map[string]any{
		"type": "mention_in_comment",
		"at":   at.UnixMilli(),
		"details": map[string]any{
			"in_post_id":      commentId,
			"mentioning_user": mentioningUser,
		},
	})

	return map[string]any{
		"notifId": notifId,
		"notif":   notif,
		"is_read": false,
	}
}
