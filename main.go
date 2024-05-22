package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocarina/gocsv"
	"os"
	"sort"
)

type APILatencies map[string]float64

type TransactionRecordWithMoneyOverTimeValue struct {
	Id                 string
	Amount             float64
	BankCountryCode    string
	latency            float64
	MoneyOverTimeValue float64
}

type TransactionRecord struct {
	Id              string  `csv:"id"`
	Amount          float64 `csv:"amount"`
	BankCountryCode string  `csv:"bank_country_code"`
}

func prioritize(transactions []*TransactionRecord, totalTime float64, apiLatencies APILatencies) ([]TransactionRecordWithMoneyOverTimeValue, float64) {
	sortedByMoneyOverTime := sortTransactionsMoneyOverTime(transactions, apiLatencies)
	println(sortedByMoneyOverTime[0])

	var prioritizedTransactions []TransactionRecordWithMoneyOverTimeValue
	var maxUSD float64
	var totalProcessedTime = 0.0

	for _, transaction := range sortedByMoneyOverTime {
		if totalProcessedTime+transaction.latency <= totalTime {
			maxUSD = maxUSD + transaction.Amount
			totalProcessedTime = transaction.latency + totalProcessedTime
			prioritizedTransactions = append(prioritizedTransactions, *transaction)
		} else {
			break
		}

	}

	return prioritizedTransactions, maxUSD
}

func sortTransactionsMoneyOverTime(transactions []*TransactionRecord, latencies APILatencies) []*TransactionRecordWithMoneyOverTimeValue {
	var transactionRecordsWithMoneyOverTimeValue []*TransactionRecordWithMoneyOverTimeValue
	for i := 0; i < len(transactions)-1; i++ {
		amount := transactions[i].Amount
		bankCountryCode := transactions[i].BankCountryCode
		latency := latencies[bankCountryCode]

		MoneyOverTimeValue := amount / latency

		transactionMoneyOverTime := TransactionRecordWithMoneyOverTimeValue{
			Id:                 transactions[i].Id,
			Amount:             amount,
			BankCountryCode:    bankCountryCode,
			latency:            latency,
			MoneyOverTimeValue: MoneyOverTimeValue,
		}

		transactionRecordsWithMoneyOverTimeValue = append(transactionRecordsWithMoneyOverTimeValue, &transactionMoneyOverTime)

	}
	sort.Slice(transactionRecordsWithMoneyOverTimeValue, func(i, j int) bool {
		return transactionRecordsWithMoneyOverTimeValue[i].MoneyOverTimeValue > transactionRecordsWithMoneyOverTimeValue[j].MoneyOverTimeValue
	})

	return transactionRecordsWithMoneyOverTimeValue
}

func createTransactionList() []*TransactionRecord {
	in, err := os.Open("transactions.csv")
	if err != nil {
		panic(err)
	}
	defer in.Close()

	var transactions []*TransactionRecord

	if err := gocsv.UnmarshalFile(in, &transactions); err != nil {
		panic(err)
	}
	return transactions
}

func loadAPILatencies(filename string) APILatencies {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var apiLatencies APILatencies
	err = json.Unmarshal(data, &apiLatencies)
	if err != nil {
		panic(err)
	}

	return apiLatencies
}

func main() {
	transactionList := createTransactionList()
	apiLatencies := loadAPILatencies("api_latencies.json")

	prioritizedTransactionsFor500ms, maxMoney500ms := prioritize(transactionList, 50, apiLatencies)
	fmt.Printf("You can do %d transactions with a maxUSD of %s when using a totalTime of 50 mil seconds \n", len(prioritizedTransactionsFor500ms), fmt.Sprintf("%.2f", maxMoney500ms))

	prioritizedTransactionsFor60ms, maxMoney60ms := prioritize(transactionList, 60, apiLatencies)
	fmt.Printf("You can do %d transactions with a maxUSD of %s when using a totalTime of 60 mil seconds \n", len(prioritizedTransactionsFor60ms), fmt.Sprintf("%.2f", maxMoney60ms))

	prioritizedTransactionsFor90ms, maxMoney90ms := prioritize(transactionList, 90, apiLatencies)
	fmt.Printf("You can do %d transactions with a maxUSD of %s when using a totalTime of 90 mil seconds \n", len(prioritizedTransactionsFor90ms), fmt.Sprintf("%.2f", maxMoney90ms))

	prioritizedTransactionsFor1000ms, maxMoney1000ms := prioritize(transactionList, 1000, apiLatencies)
	fmt.Printf("You can do %d transactions with a maxUSD of %s when using a totalTime of 1000 mil seconds \n", len(prioritizedTransactionsFor1000ms), fmt.Sprintf("%.2f", maxMoney1000ms))

}
