package main

import (
	"context"
	"fmt"

	"github.com/mbowman100/go-onfido"
)

func main() {
	ctx := context.Background()

	client, err := onfido.NewClientFromEnv()
	if err != nil {
		panic(err)
	}
	if client.Token().Prod() {
		panic("onfido token is only for production use")
	}

	applicant, err := client.CreateApplicant(ctx, onfido.Applicant{
		Email:     "rcrowe@example.co.uk",
		FirstName: "Rob",
		LastName:  "Crowe",
		Address: onfido.Address{
			BuildingNumber: "18",
			Street:         "Wind Corner",
			Town:           "Crawley",
			State:          "West Sussex",
			Postcode:       "NW9 5AB",
			Country:        "GBR",
			StartDate:      "2018-02-10",
		},
	})
	if err != nil {
		panic(err)
	}

	check, err := client.CreateCheck(ctx, onfido.CheckRequest{
		ApplicantID:           applicant.ID,
		ApplicantProvidesData: true,
		ReportNames: []string{
			string(onfido.ReportNameDocument),
			string(onfido.ReportNameIdentityEnhanced),
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Form: %+v\n", check.FormURI)
}
