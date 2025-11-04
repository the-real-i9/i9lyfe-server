package mailService

import (
	"crypto/tls"
	"fmt"
	"i9lyfe/src/helpers"
	"os"

	"gopkg.in/gomail.v2"
)

func SendMail(email string, subject string, body string) {
	user := os.Getenv("MAILING_EMAIL")
	pass := os.Getenv("MAILING_PASSWORD")

	m := gomail.NewMessage()
	m.SetHeader("From", user)
	m.SetHeader("To", email)
	m.SetHeader("Subject", fmt.Sprintf("i9lyfe - %s", subject))
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.gmail.com", 465, user, pass)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: os.Getenv("GO_ENV") != "production"}

	if err := d.DialAndSend(m); err != nil {
		helpers.LogError(err)
		return
	}
}
