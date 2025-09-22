package main

import (
	"fmt"

	"github.com/troptropcontent/qr_code_maintenance/internal/services/email"
)

type Credentials struct {
	Password string
	Username string
}

func main() {
	emailService, err := email.NewSMTPServiceGmail()
	if err != nil {
		fmt.Println("failed to instanciate email service: ", err)
	}
	err = emailService.Send("tomecrepont@gmail.com", "coucou", "Aoutch", "qr_codes/qr_1da7447b-d9ec-4fda-a002-f3420ca7b1e3.png")
	if err != nil {
		fmt.Println("failed to send email: ", err)
	}
}
