package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type fundingRatesResponse struct {
	Result struct {
		List []struct {
			FundingRate string `json:"fundingRate"`
		} `json:"list"`
	} `json:"result"`
}

func getFundingRates() ([]float64, error) {
	var fundingRates []float64

	v := url.Values{
		"category": {"inverse"},
		"symbol":   {"BTCUSD"}}

	u := url.URL{
		Scheme:   "https",
		Host:     "api.bybit.com",
		Path:     "/v5/market/funding/history",
		RawQuery: v.Encode()}

	resp, err := http.Get(u.String())
	if err != nil {
		return fundingRates, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fundingRates, err
	}

	var result fundingRatesResponse
	json.Unmarshal(body, &result)

	for _, item := range result.Result.List {
		funding, err := strconv.ParseFloat(item.FundingRate, 64)
		if err != nil {
			continue
		}

		fundingRates = append(fundingRates, funding)
	}

	return fundingRates, err
}

func main() {
	fundingRates, err := getFundingRates()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var total float64
	days := float64(len(fundingRates)) / 3

	for _, funding := range fundingRates {
		total += funding
	}

	annualised := total / (days / 364)

	fmt.Printf("Days:  %.1f\n", days)
	fmt.Printf("Total: %.2f%%\n", total*100)
	fmt.Printf("APR:   %.2f%%\n", annualised*100)
}
