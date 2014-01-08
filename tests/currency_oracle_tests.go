package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"github.com/alonsovidales/go_ml"
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

func prepare(x []float64) (res []float64) {
	// Add the bias
	res = []float64{}
	avg := 0.0
	trend := 0.0

	fmt.Println("X:", x)

	for i := 0; i < len(x) - 1; i++ {
		avg += x[i + 1]
		//res = append(res, x[i + 1])
		trend += x[i + 1] / x[i]
		fmt.Println("Trend:",  x[i + 1] / x[i])
		//fmt.Println(x[i + 1] / x[i])
	}
	//res = append(res, x[0])
	res = append(res, avg / float64(len(x) - 1))
	//fmt.Println("avg", avg / float64(len(x) - 1))
	//fmt.Println("trend", trend / float64(len(x) - 1))
	res = append(res, (trend * x[len(x) / 2]) * (trend * x[len(x) / 2]))

	//fmt.Println("Res:", res, "Avg:", avg, avg / float64(len(x) - 1), "Thrend:", trend, trend / float64(len(x) - 1))
	return
}

func prepareAll(x [][]float64) (res [][]float64) {
	res = make([][]float64, len(x))
	for i := 0; i < len(x); i++ {
		res[i] = prepare(x[i])
	}

	return
}

var test int
func getFutureBidPrice(windowAsk []float64, windowBid []float64, sectionWidth int, stepsFut int) (hip float64) {
	//window, valid := ml.Normalize(windowAsk)
	window := windowAsk
	valid := true
	if !valid {
		fmt.Println(0.0)
	} else {
		x := [][]float64{}
		y := []float64{}
		for sec := 0; sec < len(window) - sectionWidth - stepsFut; sec++ {
			section := window[sec:sec + sectionWidth]
			//fmt.Println("Last:", section[len(section) - 1], "To Predict:", windowAsk[sec + sectionWidth + stepsFut])
			prediction := windowAsk[sec + sectionWidth + stepsFut]

			x = append(x, section)
			y = append(y, prediction)
		}

		reg := &ml.Regression {
			X: prepareAll(x),
			Y: y,
			LinearReg: true,
		}
		reg.InitializeTheta()
		ml.Fmincg(reg, 0.0, 1000, false)

		lastSlice := prepare(window[len(window) - sectionWidth:])
		hip =  reg.LinearHipotesis(lastSlice)
		//fmt.Println(test, "Last:", windowBid[len(windowBid) - 1], "Hip:", hip, "Dist:", hip - windowBid[len(windowBid) - 1])
		test++
	}

	return
}

func makePredictions(sectionWidth int, windowSize int, stepsFut int, askPath string, bidPath string) {
	//fmt.Println("sectionWidth:", sectionWidth, "windowSize:", windowSize, "stepsFut:", stepsFut, "askPath:", askPath, "bidPath:", bidPath)

	askPrices := loadFile(askPath)
	bidPrices := loadFile(bidPath)
	predictions := []float64{}
	test = 0

	// Build the data to work with
	for w := 0; w <= len(askPrices) - windowSize; w++ {
		prediction := getFutureBidPrice(
			askPrices[w:w + windowSize],
			bidPrices[w:w + windowSize],
			sectionWidth,
			stepsFut)

		predictions = append(predictions, prediction)
		/*if w >= stepsFut {
			window := bidPrices[w:w + windowSize]
			fmt.Println(w - stepsFut, "Preds:", predictions[w - stepsFut], "Real:", window[len(window) - 1], "Distance:", predictions[w - stepsFut] - window[len(window) - 1])
		}*/
		fmt.Println(prediction)
	}
}

func main() {
	makePredictions(
		4, // sectionWidth
		1200, // windowSize
		5, // stepsFut -> 0 == next sample
		"usd_rate_ask.csv",
		"usd_rate_bid.csv")
}
