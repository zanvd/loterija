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
	resultNums := make(map[int]*resultNum)
	drawNum := 0

	data, err := Read()
	if err != nil {
		fmt.Println("failed to read cache:", err)
	} else {
		resultNums = data.ResultNums
		drawNum = data.DrawsNum
	}

	if data.LastVisit.Before(time.Now().UTC().Truncate(24 * time.Hour)) {
		fmt.Println("Refreshing data.")

		c := colly.NewCollector()

		crawl(c, &drawNum, resultNums)
	}

	sortedNums := make([]int, 0, len(resultNums))
	for n := range resultNums {
		sortedNums = append(sortedNums, n)
	}
	sort.SliceStable(sortedNums, func(i, j int) bool {
		return resultNums[sortedNums[i]].Normal < resultNums[sortedNums[j]].Normal
	})

	printToCmd(drawNum, sortedNums, resultNums)

	if err := Write(drawNum, resultNums); err != nil {
		fmt.Println("failed to cache:", err)
	}
}

type resultNum struct {
	Normal  int
	Special int
}

func crawl(c *colly.Collector, drawNum *int, resultNums map[int]*resultNum) {
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
		if _, ok := resultNums[num]; ok {
			resultNums[num].Normal++
		} else {
			resultNums[num] = &resultNum{
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
		if _, ok := resultNums[num]; ok {
			resultNums[num].Special++
		} else {
			resultNums[num] = &resultNum{
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

func getLowestOccurringSpecial(sortedNums map[int]*resultNum) int {
	currMinNum := 1
	currMinVal := sortedNums[currMinNum].Special
	for i, n := range sortedNums {
		if n.Special < currMinVal {
			currMinNum = i
			currMinVal = n.Special
		}
	}
	return currMinNum
}

func getHighestOccurringSpecial(sortedNums map[int]*resultNum) int {
	currHighNum := 1
	currHighVal := sortedNums[currHighNum].Special
	for i, n := range sortedNums {
		if n.Special > currHighVal {
			currHighNum = i
			currHighVal = n.Special
		}
	}
	return currHighNum
}

func getRandomNumbers() []int {
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

func printToCmd(drawNum int, sortedNums []int, resultNums map[int]*resultNum) {
	highestNormal := make([]int, 0, 7)
	lowestNormKey := 0
	lowestNormal := make([]int, 0, 7)
	fmt.Println("Total draws:", drawNum)
	fmt.Printf("Number\tNormal\tSpecial\t%%\n")
	for i, n := range sortedNums {
		if i >= len(sortedNums)-7 {
			highestNormal = append([]int{n}, highestNormal...)
		}
		if lowestNormKey < 7 {
			lowestNormal = append(lowestNormal, n)
			lowestNormKey++
		}
		fmt.Printf(
			"%v\t%v\t%v\t%.2f\n",
			n,
			resultNums[n].Normal,
			resultNums[n].Special,
			float32(resultNums[n].Normal)/float32(drawNum)*100,
		)
	}
	fmt.Println("Lowest occurring numbers:")
	fmt.Println("\t- normal:", lowestNormal)
	fmt.Println("\t- special:", getLowestOccurringSpecial(resultNums))
	fmt.Println("Highest occurring numbers:")
	fmt.Println("\t- normal:", highestNormal)
	fmt.Println("\t- special:", getHighestOccurringSpecial(resultNums))
	fmt.Println("Random numbers are:", getRandomNumbers())
	fmt.Println("Done")
}
