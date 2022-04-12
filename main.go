package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gocolly/colly"
)

func main() {
	results := make(map[int]*Result)
	draws := 0

	c := colly.NewCollector()
	c.OnResponse(func(r *colly.Response) {
		if r.StatusCode != 200 {
			fmt.Println("status code", r.StatusCode)
		}
	})
	c.OnHTML("#loto .archive-element", func(_ *colly.HTMLElement) {
		draws++
	})
	c.OnHTML("#loto .number.bg-prim", func(e *colly.HTMLElement) {
		num, err := strconv.Atoi(e.Text)
		if err != nil {
			fmt.Println(err)
			fmt.Printf("%T, %v", num, num)
		}
		if _, ok := results[num]; ok {
			results[num].Normal++
		} else {
			results[num] = &Result{
				Normal:  1,
				Special: 0,
			}
		}
	})
	c.OnHTML("#loto .number.additional", func(e *colly.HTMLElement) {
		num, err := strconv.Atoi(e.Text)
		if err != nil {
			fmt.Println(err)
			fmt.Printf("%T, %v", num, num)
		}
		if _, ok := results[num]; ok {
			results[num].Special++
		} else {
			results[num] = &Result{
				Normal:  0,
				Special: 1,
			}
		}
	})

	for y := 1991; y <= time.Now().Year(); y++ {
		fmt.Println("Year", y)
		url := "https://www.loterija.si/loto/rezultati?year=" + strconv.Itoa(y) + "&selectedGame=loto&ajax=.archive-dynamic"
		err := c.Visit(url)
		if err != nil {
			fmt.Println("Failed to visit", url)
			fmt.Println(err)
		}
	}

	fmt.Println("Total draws:", draws)
	fmt.Printf("Number\tNormal\tSpecial\t%%\n")
	for num, result := range results {
		fmt.Printf("%v\t%v\t%v\t%f\n", num, result.Normal, result.Special, float32(result.Normal)/float32(draws)*100)
	}
	fmt.Println("Done")
}

type Result struct {
	Normal  int
	Special int
}
