package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

type Reading struct {
	County string
	Counts []int64
}

func main() {
	resp, err := http.Get("https://raw.githubusercontent.com/CSSEGISandData/COVID-19/master/csse_covid_19_data/csse_covid_19_time_series/time_series_covid19_confirmed_US.csv")
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	c := cvs2map(resp.Body)

	for state, countys := range c {
		for _, reading := range countys {
			decline := daysDecline(reading.Counts)
			if (decline > 0) && (decline != len(reading.Counts)) {
				fmt.Printf("%02d: %s - %s\n", decline, reading.County, state)
			}
		}
	}
}

func toInts(values []string) []int64 {
	var a []int64
	for _, value := range values {
		f, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		a = append(a, f)
	}

	return a
}

func cvs2map(data io.Reader) map[string][]Reading {
	covid := make(map[string][]Reading)
	r := csv.NewReader(data)

	// first line is a header line, throw it away
	r.Read()

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// skip readings that don't have a LAT/LONG
		if record[9] == "0.0" {
			continue
		}

		covid[record[6]] = append(covid[record[6]], Reading{County: record[5], Counts: toInts(record[11:])})
	}

	return covid
}

func daysDecline(l []int64) int {
	var last int64
	last = 0

	var trend int
	for _, count := range l {
		if count-last < 1 {
			trend = trend + 1
		} else {
			trend = 0
		}
		last = count
	}

	return trend
}
