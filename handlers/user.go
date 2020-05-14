package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"

	"../managers/database"
)

func (t *MethodInterface) VerifyUserEmail(w http.ResponseWriter, args map[string]interface{}) (valid bool) {
	fields := []string{"user_id"}
	userId, err := database.Read("admin", "vr_user", fields, "", args)

	if err != nil {
		if _, err := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); err != nil {
			log.Fatal(err)
		}
		return
	}

	if userId != nil {
		auth := smtp.PlainAuth("", os.Getenv("VR_SMTP_USR"),
			os.Getenv("VR_SMTP_PWD"), os.Getenv("VR_SMTP_HOST"))

		// Connect to the server, authenticate, set the sender and recipient,
		// and send the email all in one step.
		to := []string{"dslipak@gmail.com"}
		msg := []byte("To: dslipak@gmail.com\r\n" +
			"Subject: discount Gophers!\r\n" +
			"\r\n" +
			"This is the email body.\r\n")
		err := smtp.SendMail(fmt.Sprintf("%s:%s", os.Getenv("VR_SMTP_HOST"),
			os.Getenv("VR_SMTP_PORT")), auth, "dslipak@gmail.com", to, msg)
		if err != nil {
			log.Fatal(err)
		}
	}

	return userId != nil
}
