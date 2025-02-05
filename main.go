package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

func main() {
	fName := "/tmp/cigardsdb_nobelgo.json"
	f, err := os.OpenFile(fName, os.O_CREATE|os.O_RDWR, 0774)
	if err != nil {
		log.Printf("could not open file: %w", err)
		return
	}
	defer func() { _ = f.Close() }()

	data, err := ExtractNobelgo(http.DefaultClient)
	if err != nil {
		log.Println(err)
		return
	}

	if err = json.NewEncoder(f).Encode(data); err != nil {
		log.Printf("could write data to %s: %w", fName, err)
	}
}

type httpClient interface {
	Get(url string) (resp *http.Response, err error)
}

type Record struct {
	Name           string   `json:"name"`
	URL            string   `json:"url"`
	Manufacturer   string   `json:"manufacturer"`
	OriginCountry  string   `json:"manufacturerCountry"`
	Collection     string   `json:"collection"`
	Format         string   `json:"format"`
	Form           string   `json:"form"`
	WrapperCountry []string `json:"wrapperCountry"`
	FillerCountry  []string `json:"fillingCountry"`
	BinderCountry  []string `json:"binderCountry"`
	Strength       string   `json:"strength"`
	Price          float64  `json:"price"`
	Diameter       float64  `json:"diameter"`
	Circumference  float64  `json:"circumference"`
	Length         float64  `json:"length"`
	Aromas         []string `json:"aromas"`
}

func ExtractNobelgo(c httpClient) (o []Record, err error) {
	page := 1
	for {
		// filters:
		//	- the gauge: between 50 and 60
		//  - only in stock
		// sorted by price in ascending order
		baseURL := fmt.Sprintf("https://www.noblego.de/zigarren/?cig_gauge%5B0%5D=218&cig_gauge%5B1%5D=217&cig_gauge%5B2%5D=216&"+"cig_gauge%5B3%5D=215&cig_gauge%5B4%5D=214&cig_gauge%5B5%5D=213&cig_gauge%5B6%5D=212&cig_gauge%5B7%5D=1025&cig_gauge%5B8%5D=211&cig_gauge%5B9%5D=210&dir=asc&in_stock=1&limit=96&order=price&p=%d", page)
		var er error
		if resp, er := c.Get(baseURL); er == nil {
			if data, er := io.ReadAll(resp.Body); er == nil {
				func() { _ = resp.Body.Close() }()
				if v, found, er := readPage(data, c); er == nil {
					if found {
						o = append(o, v...)
					} else {
						break
					}
				}
			}
		}
		if er != nil {
			err = errors.Join(err, fmt.Errorf("error reading page %d: %w", page, er))
		}
		page++
	}
	return o, err
}

func readPage(v []byte, c httpClient) (o []Record, found bool, err error) {
	o = readList(v)
	var mu = new(sync.Mutex)
	for i, el := range o {
		i := i
		el := el
		go func() {
			var (
				er   error
				resp *http.Response
				data []byte
			)
			if resp, er = c.Get(el.URL); er == nil {
				defer func() { _ = resp.Body.Close() }()
				if data, er = io.ReadAll(resp.Body); er == nil {
					func() { _ = resp.Body.Close() }()
					if er = readDetails(data, &el); er == nil {
						o[i] = el
					}
				}
			}
			if er != nil {
				mu.Lock()
				err = errors.Join(err, fmt.Errorf("error reading details using %s: %w", el.URL, er))
				mu.Unlock()
			}
		}()
	}
	return o, len(o) > 0, err
}

func readList(v []byte) []Record {
	panic("todo")
}

func readDetails(data []byte, o *Record) error {
	panic("todo")
}
