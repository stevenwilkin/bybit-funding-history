package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type fundingRate struct {
	FundingRate          string `json:"fundingRate"`
	FundingRateTimestamp string `json:"fundingRateTimestamp"`
}

type fundingRatesResponse struct {
	Result struct {
		List []fundingRate `json:"list"`
	} `json:"result"`
}

func getPage(startTime, endTime int64) ([]fundingRate, error) {
	v := url.Values{
		"category":  {"inverse"},
		"symbol":    {"BTCUSD"},
		"startTime": {fmt.Sprintf("%d", startTime)},
		"endTime":   {fmt.Sprintf("%d", endTime)}}

	u := url.URL{
		Scheme:   "https",
		Host:     "api.bybit.com",
		Path:     "/v5/market/funding/history",
		RawQuery: v.Encode()}

	resp, err := http.Get(u.String())
	if err != nil {
		return []fundingRate{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []fundingRate{}, err
	}

	var result fundingRatesResponse
	json.Unmarshal(body, &result)

	return result.Result.List, nil
}

func getFundingRates(initialTime int64) ([]float64, error) {
	var (
		results      []float64
		fundingRates []fundingRate
		err          error
	)

	startTime := initialTime
	endTime := time.Now().UnixNano() / int64(time.Millisecond)

	for {
		fundingRates, err = getPage(startTime, endTime)
		if err != nil {
			return []float64{}, err
		}

		for _, item := range fundingRates {
			funding, err := strconv.ParseFloat(item.FundingRate, 64)
			if err != nil {
				continue
			}

			results = append(results, funding)
		}

		if len(fundingRates) == 200 {
			// results are paged in reverse chronological order
			oldestItem := fundingRates[len(fundingRates)-1]
			ts, _ := strconv.Atoi(oldestItem.FundingRateTimestamp)
			endTime = int64(ts) - 1
		} else {
			break
		}
	}

	return results, err
}

func startTime() int64 {
	var days int

	if len(os.Args) == 2 {
		days, _ = strconv.Atoi(os.Args[1])
	}

	if days == 0 {
		days = 30
	}

	ts := time.Now().AddDate(0, 0, days*-1)

	return ts.UnixNano() / int64(time.Millisecond)
}

func main() {
	fundingRates, err := getFundingRates(startTime())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var total float64
	days := float64(len(fundingRates)) / 3

	for _, funding := range fundingRates {
		total += funding
	}

	annualised := total / (days / 365)

	fmt.Printf("Days:  %.1f\n", days)
	fmt.Printf("Total: %.2f%%\n", total*100)
	fmt.Printf("APR:   %.2f%%\n", annualised*100)
}
