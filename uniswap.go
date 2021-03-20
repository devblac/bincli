package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	uniswapAliases = map[string]string{
		"APYS":   "0xf7413489c474ca4399eee604716c72879eea3615",
		"SYNC":   "0xb6ff96b8a8d214544ca0dbc9b33f7ad6503efd32",
		"ROOM":   "0xad4f86a25bbc20ffb751f2fac312a0b4d8f88c64",
		"EROWAN": "0x07bac35846e5ed502aa91adf6a9e7aa210f2dcbe",
		"INV":    "0x41d5d79431a913c4ae7d69a668ecdfe5ff9dfb68",
		"MARK":   "0x67c597624b17b16fb77959217360b7cd18284253",
	}
)

type tokenDayData struct {
	PriceUSD string `json:"priceUSD"`
}

type tokenDayDatas struct {
	TokenDayDatas []tokenDayData `json:"tokenDayDatas"`
}

type response struct {
	Data tokenDayDatas `json:"data"`
}

func uniswapTicker() {
	if len(os.Args) < 3 {
		usage()
	}
	symbol := os.Args[2]
	contractAddress := symbol
	if uniswapAliases[symbol] != "" {
		contractAddress = uniswapAliases[symbol]
	}

	price, err := requestUniswapTicker(contractAddress)
	if err != nil {
		log.Printf("Error getting ticker price for %v (because %v)", contractAddress, err)
		usage()
	}
	fmt.Println(price)
}

func requestUniswapTicker(contractAddress string) (float64, error) {
	url := "https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v2"

	var jsonStr = []byte(`{"query": "{tokenDayDatas(first: 1, orderBy: date, orderDirection: desc, where: { token: \"` + contractAddress + `\"}) {priceUSD } }"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	responseData := response{}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return 0, err
	}
	strPrice := responseData.Data.TokenDayDatas[0].PriceUSD
	price, err := strconv.ParseFloat(strPrice, 64)
	if err != nil {
		return 0, err
	}
	return price, nil
}

func isUniswapConditionMet(contractAddress string, comparator string, target float64) (float64, bool, error) {
	price, err := requestUniswapTicker(contractAddress)
	if err != nil {
		return price, false, err
	}
	return isConditionMet(price, comparator, target)
}

func uniswapAlert() {
	if len(os.Args) < 5 {
		usage()
	}
	var (
		symbol      = os.Args[2]
		comparator  = os.Args[3]
		targetStr   = os.Args[4]
		target, err = strconv.ParseFloat(targetStr, 64)
	)
	if err != nil {
		log.Fatal(err)
	}

	contractAddress := symbol
	if uniswapAliases[symbol] != "" {
		contractAddress = uniswapAliases[symbol]
	}

	for {
		price, isConditionMet, err := isUniswapConditionMet(contractAddress, comparator, target)
		if err != nil {
			log.Printf("Error getting ticker price for %v (because %v)", symbol, err)
		} else {
			log.Printf("%v %v %v where %v = %v...\n",
				symbol,
				comparator,
				targetStr,
				symbol,
				price,
			)
		}
		if isConditionMet {
			break
		}
		time.Sleep(10 * time.Second)
	}
	log.Println("Condition met!")
}
