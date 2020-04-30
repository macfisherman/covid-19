package main

import (
	"encoding/csv"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-echarts/go-echarts/charts"
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

	state := os.Args[1]
	spew.Dump(c[state])

	line := charts.NewLine()
	line.SetGlobalOptions(charts.TitleOpts{Title: "Covid-19 Cases"},
		charts.LegendOpts{Type: "scroll", Orient: "horizontal", Bottom: "0"})
	labels := dayLabels(len(c[state][0].Counts))
	line.AddXAxis(labels)

	for _, reading := range c[state] {
		line.AddYAxis(reading.County, reading.Counts)
		log.Println("Added " + reading.County)
		log.Printf("Downward trend: %d\n", daysDecline(reading.Counts))
	}

	f, err := os.Create(state + ".html")
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	line.Render(f)
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

func dayLabels(l int) []int64 {
	var a []int64
	for i := 0; i < l; i++ {
		a = append(a, int64(i))
	}

	return a
}

func daysDecline(l []int64) int {
	var last int64
	last = 1000000000000

	var trend int
	for _, count := range l {
		if count < last {
			trend = trend + 1
			last = count
		} else {
			trend = 0
		}
	}

	return trend
}
