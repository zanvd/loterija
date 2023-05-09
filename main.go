package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

func main() {
	results := make(map[int]*result)
	drawNum := 0

	c := colly.NewCollector()

	crawl(c, &drawNum, results)

	nums := make([]int, 0, len(results))
	for n := range results {
		nums = append(nums, n)
	}
	sort.SliceStable(nums, func(i, j int) bool {
		return results[nums[i]].Normal < results[nums[j]].Normal
	})

	printToCmd(drawNum, nums, results)
}

type result struct {
	Normal  int
	Special int
}

func crawl(c *colly.Collector, drawNum *int, results map[int]*result) {
	c.OnResponse(func(r *colly.Response) {
		if r.StatusCode != 200 {
			fmt.Println("status code", r.StatusCode)
		}
	})
	c.OnHTML("#loto .archive-element", func(_ *colly.HTMLElement) {
		*drawNum++
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
			results[num] = &result{
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
		if _, ok := results[num]; ok {
			results[num].Special++
		} else {
			results[num] = &result{
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
}

func getLowestOccurringSpecial(nums map[int]*result) int {
	currMinNum := 1
	currMinVal := nums[currMinNum].Special
	for i, n := range nums {
		if n.Special < currMinVal {
			currMinNum = i
			currMinVal = n.Special
		}
	}
	return currMinNum
}

func getHighestOccurringSpecial(nums map[int]*result) int {
	currHighNum := 1
	currHighVal := nums[currHighNum].Special
	for i, n := range nums {
		if n.Special > currHighVal {
			currHighNum = i
			currHighVal = n.Special
		}
	}
	return currHighNum
}

func getRandomNumbers() []int {
	rand.Seed(time.Now().UnixNano())
	nums := make([]int, 0, 7)
	for i := 0; i < 7; {
		num := rand.Intn(40) + 1
		exists := false
		for _, n := range nums {
			if n == num {
				exists = true
			}
		}
		if !exists {
			nums = append(nums, num)
			i++
		}
	}
	return nums
}

func getURL(page int, year int) string {
	return "https://www.loterija.si/loto/rezultati?selectedGame=loto&ajax=.archive-dynamic" +
		"&page=" + strconv.Itoa(page) +
		"&year=" + strconv.Itoa(year)
}

func printToCmd(drawNum int, nums []int, results map[int]*result) {
	highestNormal := make([]int, 0, 7)
	lowestNormKey := 0
	lowestNormal := make([]int, 0, 7)
	fmt.Println("Total draws:", drawNum)
	fmt.Printf("Number\tNormal\tSpecial\t%%\n")
	for i, n := range nums {
		if i >= len(nums)-7 {
			highestNormal = append([]int{n}, highestNormal...)
		}
		if lowestNormKey < 7 {
			lowestNormal = append(lowestNormal, n)
			lowestNormKey++
		}
		fmt.Printf(
			"%v\t%v\t%v\t%.2f\n",
			n,
			results[n].Normal,
			results[n].Special,
			float32(results[n].Normal)/float32(drawNum)*100,
		)
	}
	fmt.Println("Lowest occurring numbers:")
	fmt.Println("\t- normal:", lowestNormal)
	fmt.Println("\t- special:", getLowestOccurringSpecial(results))
	fmt.Println("Highest occurring numbers:")
	fmt.Println("\t- normal:", highestNormal)
	fmt.Println("\t- special:", getHighestOccurringSpecial(results))
	fmt.Println("Random numbers are:", getRandomNumbers())
	fmt.Println("Done")
}
