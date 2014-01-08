package main

import (
	"fmt"
	"math"
	"io/ioutil"
	"strconv"
	"strings"
)

func loadFile(filePath string) (data []float64) {
	strInfo, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	trainingData := strings.Split(string(strInfo), "\n")
	for _, line := range trainingData {
		if line == "" {
			break
		}
		num, _ := strconv.ParseFloat(line, 64)
		data = append(data, num)
	}

	return
}

func getFutureBidPrice(section []float64, stepsFut int) (r float64, thrend float64) {
	//fmt.Println(section)
	thrend = 0.0
	for i := 0; i < len(section) - 1; i++ {
		//fmt.Println("Var:", section[i + 1] / section[i])
		thrend += section[i + 1] / section[i]
	}
	thrend /= float64(len(section) - 1)

	r = section[len(section) - 1]
	for i := 0; i < stepsFut; i++ {
		r *= thrend
	}
	//fmt.Println("Trend", thrend)

	return
}

func isThrendValid(validThrends []float64, unvalidThrends []float64, thrend float64) bool {
	if len(validThrends) == 0 || len(unvalidThrends) == 0 {
		return true
	}

	costValid := 0.0
	for _, validThrend := range validThrends {
		costValid += math.Pow(thrend - validThrend, 2)
	}
	costValid /= float64(len(validThrends))

	costUnvalid := 0.0
	for _, unvalidThrend := range unvalidThrends {
		costUnvalid += math.Pow(thrend - unvalidThrend, 2)
	}
	costUnvalid /= float64(len(unvalidThrends))

	return costValid <= costUnvalid
}

func makePredictions(sectionWidth int, stepsFut int, margin float64, packSize float64, askPath string, bidPath string) {
	//fmt.Println("sectionWidth:", sectionWidth, "windowSize:", windowSize, "stepsFut:", stepsFut, "askPath:", askPath, "bidPath:", bidPath)

	askPrices := loadFile(askPath)
	bidPrices := loadFile(bidPath)
	total := 100.0
	inverted := 0.0
	boughtAt := 0.0
	threndAt := 0.0
	validThrends := []float64{}
	unvalidThrends := []float64{}

	// Build the data to work with
	for w := 0; w <= len(askPrices) - sectionWidth; w++ {
		section := bidPrices[w:w + sectionWidth]
		futBid, thrend := getFutureBidPrice(section, stepsFut)
		currentBid := bidPrices[w + sectionWidth - 1]
		currentAsk := askPrices[w + sectionWidth - 1]

		// Buy
		if futBid > currentAsk && inverted == 0 && isThrendValid(validThrends, unvalidThrends, thrend) {
			total -= packSize
			inverted = packSize / currentAsk
			//fmt.Println(section)
			//fmt.Println(askPrices[w:w + sectionWidth])
			fmt.Println(w + sectionWidth, "Buy:", currentAsk, total)
			boughtAt = currentAsk
			threndAt = thrend
		}

		// Sell
		if inverted != 0 && futBid < currentBid {
			total += inverted * currentBid
			inverted = 0
			fmt.Println(w + sectionWidth, "Sell:", currentBid, total, (currentBid - boughtAt))

			if (currentBid - boughtAt) > 0 {
				fmt.Println("Is Valid!!!")
				validThrends = append(validThrends, threndAt)
			} else {
				fmt.Println("Is Invalid!!!")
				unvalidThrends = append(unvalidThrends, threndAt)
			}
		}
	}

	fmt.Println("Total:", total)
}

func main() {
	makePredictions(
		4, // sectionWidth
		4, // stepsFut -> 0 == next sample
		0.00005, // margin
		1.0, // Pack size
		"usd_rate_ask.csv",
		"usd_rate_bid.csv")
}
