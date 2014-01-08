package curr_oracle

import (
	"fmt"
	"strings"
	"time"
	"os"
	"math"
	"strconv"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"sync"
)

const (
	CLR_0 = "\x1b[30;1m"
	CLR_R = "\x1b[31;1m"
	CLR_G = "\x1b[32;1m"
	CLR_Y = "\x1b[33;1m"
	CLR_B = "\x1b[34;1m"
	CLR_M = "\x1b[35;1m"
	CLR_C = "\x1b[36;1m"
	CLR_W = "\x1b[37;1m"
	CLR_N = "\x1b[0m"

	FEEDS_URL = "http://api-sandbox.oanda.com/v1/quote?instruments="

	VERBOSE = false
	PRODUCTION = false

	THRENDS_TO_CONSIDER = 5
	MAX_BENEFIT_TO_CONSIDERER = 10

	INCR_MARGIN = 1.0
)

type oracleStruc struct {
	baserCurr string
	currencies []string
	samplesToStudy int
	samplePeriod time.Duration
	futurePeriod int
	units float64
	unitsByOp float64
	internalClock int64
	mutex sync.Mutex
	account *accountStruct

	samplesBid map[string][]float64
	samplesAsk map[string][]float64
}

type feedStruc struct {
	Instrument string `json:"instrument"`
	Time string `json:"time"`
	Bid float64 `json:"bid"`
	Ask float64 `json:"ask"`
}

func (oracle *oracleStruc) collectSamples() {
	var feeds map[string][]feedStruc

	fa, _ := os.OpenFile("usd_rate_ask.csv", os.O_APPEND|os.O_WRONLY, 0600)
	defer fa.Close()
	fb, _ := os.OpenFile("usd_rate_bid.csv", os.O_APPEND|os.O_WRONLY, 0600)
	defer fb.Close()

	currenciesList := make([]string, len(oracle.currencies))
	oracle.samplesBid = make(map[string][]float64)
	oracle.samplesAsk = make(map[string][]float64)
	lasCurrPrice := make(map[string]string)
	for i := 0; i < len(oracle.currencies); i++ {
		currenciesList[i] = fmt.Sprintf("%s_%s", oracle.baserCurr, oracle.currencies[i])

		oracle.samplesBid[oracle.currencies[i]] = []float64{}
		oracle.samplesAsk[oracle.currencies[i]] = []float64{}

		lasCurrPrice[oracle.currencies[i]] = ""
	}
	feedsUrl := FEEDS_URL + strings.Join(currenciesList, "%2C")
	fmt.Println("Parsing feeds from:", feedsUrl)

	feedsLoop: for {
		initTime := time.Now().UnixNano()

		resp, err := http.Get(feedsUrl)
		if err != nil {
			fmt.Println("Error:", err)
			// We found a problem reading the feedss, loock all the processes
			continue feedsLoop
		}

		body, err:= ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error:", err)
			// We found a problem reading the feedss, loock all the processes
			continue feedsLoop
		}

		err = json.Unmarshal(body, &feeds)
		if err != nil {
			fmt.Println("Error:", err)
			// We found a problem reading the feedss, loock all the processes
			continue feedsLoop
		}

		oracle.mutex.Lock()
		for _, feed := range feeds["prices"] {
			curr := feed.Instrument[len(oracle.baserCurr) + 1:]

			if lasCurrPrice[curr] != feed.Time {
				lasCurrPrice[curr] = feed.Time

				fmt.Println(oracle.internalClock, "Las Price:", curr, feed.Bid)

				oracle.samplesBid[curr] = append(oracle.samplesBid[curr], feed.Bid)
				oracle.samplesAsk[curr] = append(oracle.samplesAsk[curr], feed.Ask)

				if len(oracle.samplesBid[curr]) > oracle.samplesToStudy {
					oracle.samplesBid[curr] = oracle.samplesBid[curr][1:]
					oracle.samplesAsk[curr] = oracle.samplesAsk[curr][1:]
				}

				if !PRODUCTION && curr == "EUR" {
					fb.WriteString(strconv.FormatFloat(feed.Bid, 'e', -1, 64) + "\n")
					fa.WriteString(strconv.FormatFloat(feed.Ask, 'e', -1, 64) + "\n")
				}
			}
		}
		oracle.mutex.Unlock()
		oracle.internalClock++

		time.Sleep((oracle.samplePeriod * time.Millisecond) - time.Duration((time.Now().UnixNano() - initTime) / 1000000))
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

func isThrendValid(validThrends []float64, invalidThrends []float64, thrend float64) bool {
	if len(validThrends) == 0 || len(invalidThrends) == 0 {
		return true
	}

	costValid := 0.0
	for _, validThrend := range validThrends {
		costValid += math.Pow(thrend - validThrend, 2)
	}
	costValid /= float64(len(validThrends))

	costInvalid := 0.0
	for _, invalidThrend := range invalidThrends {
		costInvalid += math.Pow(thrend - invalidThrend, 2)
	}
	costInvalid /= float64(len(invalidThrends))

	return costValid <= costInvalid
}

func sumArr(x []float64) (r float64) {
	for i := 0; i < len(x); i++ {
		r += x[i]
	}
	return
}

func (oracle *oracleStruc) makePredictions(curr string, futurePeriod int, samplesToStudy int) {
	var boughtAt, threndAt float64

	inverted := 0.0
	validThrends := []float64{}
	invalidThrends := []float64{}
	unitsByOp := 1.0
	exangeAt := 0.0
	benefit := []float64{}

	for {
		if len(oracle.samplesBid[curr]) < samplesToStudy {
			fmt.Println("Waiting...", samplesToStudy, futurePeriod, curr, "current size:", len(oracle.samplesBid[curr]))
			time.Sleep(1000 * time.Millisecond)
		} else {
			// Make a copy of the objects in order to avoid problems when update the rates
			oracle.mutex.Lock()
			samplesBid := make([]float64, samplesToStudy)
			copy(samplesBid, oracle.samplesBid[curr][len(oracle.samplesBid) - samplesToStudy:])
			samplesAsk := make([]float64, samplesToStudy)
			copy(samplesAsk, oracle.samplesAsk[curr][len(oracle.samplesBid) - samplesToStudy:])
			oracle.mutex.Unlock()

			fmt.Println("To Study---->>>>>", samplesToStudy, samplesBid)
			fmt.Println("To Study---->>>>>", samplesToStudy, samplesAsk)

			futBid, thrend := getFutureBidPrice(samplesBid, futurePeriod)
			currentBid := samplesBid[len(samplesBid) - 1]
			currentAsk := samplesAsk[len(samplesAsk) - 1]

			/*if inverted != 0 {
				fmt.Println("Thrend:", thrend)
			}*/

			if sumArr(benefit) > 0 {
				unitsByOp = oracle.unitsByOp
			} else {
				unitsByOp = 1
			}

			// Buy
			if futBid > (currentAsk * INCR_MARGIN) && inverted == 0 && isThrendValid(validThrends, invalidThrends, thrend) {
				order, err := oracle.account.placeOrder(
					fmt.Sprintf("%s_%s", oracle.baserCurr, curr),
					unitsByOp,
					"buy")

				if err != nil {
					fmt.Println("Error placing order:", err)
				} else {
					oracle.units -= unitsByOp
					inverted = unitsByOp / order.Price
					boughtAt = order.Price
					threndAt = thrend
					exangeAt = thrend
					fmt.Println(CLR_R, oracle.internalClock, "Buy:", curr, samplesToStudy, futurePeriod, order.Price, oracle.units, CLR_N)
				}
			}

			// Sell
			if inverted != 0 && futBid < currentBid {
				order, err := oracle.account.placeOrder(
					fmt.Sprintf("%s_%s", oracle.baserCurr, curr),
					unitsByOp,
					"sell")

				if err != nil {
					fmt.Println("Error placing order:", err)
				} else {
					oracle.units += inverted * order.Price
					inverted = 0
					benefit = append(benefit, order.Price - boughtAt)
					if len(benefit) > MAX_BENEFIT_TO_CONSIDERER {
						benefit = benefit[1:]
					}
					fmt.Println(CLR_G, oracle.internalClock, "Sell:", curr, samplesToStudy, futurePeriod, order.Price, oracle.units, "Diff:", order.Price - boughtAt, "Thernd at:", threndAt, "Benefit:", sumArr(benefit), CLR_N)
				}

				if (currentBid - boughtAt) > 0 {
					validThrends = append(validThrends, exangeAt)
					if len(validThrends) > THRENDS_TO_CONSIDER {
						validThrends = validThrends[1:]
					}
				} else {
					invalidThrends = append(invalidThrends, exangeAt)
					if len(invalidThrends) > THRENDS_TO_CONSIDER {
						invalidThrends = invalidThrends[1:]
					}
				}
				fmt.Println("Valid thrends:", curr, samplesToStudy, futurePeriod, validThrends)
				fmt.Println("Invalid thrends:", curr, samplesToStudy, futurePeriod, invalidThrends)
			}

			time.Sleep((oracle.samplePeriod / 3) * time.Millisecond)
		}
	}
}

func Start(currencies []string, samplesToStudy int, futurePeriod int, samplePeriod time.Duration, unitsByOp float64) (oracle *oracleStruc, err error) {
	account := fakeAccount(true)
	fmt.Println("New account created:", account.AccountId, "Pass:", account.Pass)

	oracle = &oracleStruc {
		baserCurr: account.AccountCurrency,
		currencies: currencies,
		samplePeriod: samplePeriod,
		samplesToStudy: samplesToStudy,
		futurePeriod: futurePeriod,
		units: account.Balance,
		unitsByOp: unitsByOp,
		internalClock: 0,
		account: account,
	}

	go oracle.collectSamples()
	for _, curr := range(currencies) {
		for f := 1; f <= futurePeriod; f++ {
			for s := 1; s <= samplesToStudy; s++ {
				go oracle.makePredictions(curr, f, s)
			}
		}
	}

	//oracle.makePredictions("EUR")
	for {
		time.Sleep(1000000000 * time.Millisecond)
	}
	return
}
