package main

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type crawlRes struct {
	drawnNums  drawnNums
	mu         sync.Mutex
	numOfDraws int
}

type drawnNum struct {
	Normal  int
	Special int
}

type drawnNums map[int]drawnNum

func crawl(c *colly.Collector) crawlRes {
	res := crawlRes{
		drawnNums: make(drawnNums),
	}

	c.OnResponse(func(r *colly.Response) {
		if r.StatusCode != 200 {
			fmt.Println("status code", r.StatusCode)
		}
	})
	c.OnHTML("#loto .archive-element", func(_ *colly.HTMLElement) {
		res.numOfDraws++
	})
	c.OnHTML("#loto .number.bg-prim", func(e *colly.HTMLElement) {
		num, err := strconv.Atoi(e.Text)
		if err != nil {
			fmt.Println(err)
			fmt.Printf("%T, %v", num, num)
		}

		res.mu.Lock()
		defer res.mu.Unlock()

		if dn, ok := res.drawnNums[num]; ok {
			dn.Normal++
			res.drawnNums[num] = dn
		} else {
			res.drawnNums[num] = drawnNum{
				Normal:  1,
				Special: 0,
			}
		}
	})
	c.OnHTML("#loto .number.additional", func(e *colly.HTMLElement) {
		num, err := strconv.Atoi(strings.TrimSpace(strings.ReplaceAll(e.Text, "Dodatna številka", "")))
		if err != nil {
			fmt.Println(err)
			fmt.Printf("%T, %v", num, num)
		}

		res.mu.Lock()
		defer res.mu.Unlock()

		if dn, ok := res.drawnNums[num]; ok {
			dn.Special++
			res.drawnNums[num] = dn
		} else {
			res.drawnNums[num] = drawnNum{
				Normal:  0,
				Special: 1,
			}
		}
	})
	c.OnHTML("a.pagination-arrow.right", func(e *colly.HTMLElement) {
		// Some endpoints include empty next page URLs and since there's no year with more than 2 pages,
		// it's hardcoded atm.
		/*regPage := regexp.MustCompile(".*page=(\\d+).*")
		matchesPage := regPage.FindStringSubmatch(e.Attr("href"))
		p, err := strconv.Atoi(matchesPage[1])
		if err != nil {
			fmt.Printf("failed to get the page number from %q: %s", e.Attr("href"), err)
		}*/
		p := 2

		regYear := regexp.MustCompile(".*year=(\\d+).*")
		matchesYear := regYear.FindStringSubmatch(e.Attr("href"))
		y, err := strconv.Atoi(matchesYear[1])
		if err != nil {
			fmt.Printf("failed to get the year from %q: %s", e.Attr("href"), err)
		}

		u := getURL(p, y)
		if u != e.Request.URL.String() {
			err = e.Request.Visit(u)
			if err != nil {
				fmt.Println("Failed to visit", u)
				fmt.Println(err)
			}
		}
	})

	var wg sync.WaitGroup
	for y := 1991; y <= time.Now().Year(); y++ {
		wg.Add(1)

		go func(y int) {
			defer wg.Done()

			url := getURL(1, y)
			err := c.Visit(url)
			if err != nil {
				fmt.Println("Failed to visit", url)
				fmt.Println(err)
			}
		}(y)
	}
	wg.Wait()

	return res
}

func getURL(page int, year int) string {
	return "https://www.loterija.si/loto/rezultati?selectedGame=loto&ajax=.archive-dynamic" +
		"&page=" + strconv.Itoa(page) +
		"&year=" + strconv.Itoa(year)
}
