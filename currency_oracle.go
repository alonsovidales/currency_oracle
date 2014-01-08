package main

import (
	"fmt"
	"curr_oracle"
)

func main() {
	_, err := curr_oracle.Start(
		[]string{
			//"USD", // US Dollar
			"EUR", // Euro
			"GBP", // British Pound
			"AUD", // Australian Dollar
			"CAD", // Canadian Dollar
			"CZK", // Czech Koruna
			"JPY", // Japanese Yen
			"NOK", // Norwegian Kroner
			"HUF", // Hungarian Forint
			"DKK", // Danish Krone
			"PLN", // Polish Zloti
			"SEK", // swedish krona
			"CHF", // Swiss Franc
			"ZAR", // South African Rand
			"MXN", // Mexican peso

			//"SAR", // Saudi Arabian Riyal
			//"HKD", // Hong Kong Dollar
			//"SGD", // Singapore dollar
			//"INR", // Indian rupee
			//"TWD", // Taiwan New Dollar
			//"THB", // Thai Baht
			//"TRY", // Turkish Lira
			//"CNY", // Chinese Yuan Renminbi

			//"FJD", // Fiji Dollar
			//"EGP", // Egyptian Pound
			//"ECS", // Ecuador Sucre
			//"COP", // Colombian Peso
			//"CLP", // Chilean Peso
			//"XAU", // Gold (oz.)
			//"XAG", // Silver (oz.)
			//"XPD", // Palladium (oz.)
			//"XPT", // Platinum (oz.)
			//"ARS", // Argentine Peso
			//"RUB", // Russian Rublo
			//"BDT", // Bangladeshi Taka

		},
		5,  // samplesToStudy
		2,    // futurePeriod
		500,  // samplePeriod
		500) // unitsByOp

	if err != nil {
		fmt.Println(err)
	}
}
