package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/gocolly/colly/v2"
)

func main() {
	countedNums := make(drawnNums)
	numOfDraws := 0

	data, err := read()
	if err != nil {
		fmt.Println("failed to read cache:", err)
	} else {
		countedNums = data.DrawnNums
		numOfDraws = data.NumOfDraws
	}

	if data.LastVisit.Before(time.Now().UTC().Truncate(24 * time.Hour)) {
		fmt.Println("Refreshing data.")

		c := colly.NewCollector()

		res := crawl(c)
		countedNums = res.drawnNums
		numOfDraws = res.numOfDraws
	}

	sortedNums := make([]int, 0, len(countedNums))
	for n := range countedNums {
		sortedNums = append(sortedNums, n)
	}
	sort.SliceStable(sortedNums, func(i, j int) bool {
		return countedNums[sortedNums[i]].Normal < countedNums[sortedNums[j]].Normal
	})

	if err := write(numOfDraws, countedNums); err != nil {
		fmt.Println("failed to cache:", err)
	}

	printToCmd(numOfDraws, sortedNums, countedNums)
}

func getLowestOccurringSpecial(sortedNums drawnNums) int {
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

func getHighestOccurringSpecial(sortedNums drawnNums) int {
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

func printToCmd(numOfDraws int, sortedNums []int, resultNums drawnNums) {
	highestNormal := make([]int, 0, 7)
	lowestNormKey := 0
	lowestNormal := make([]int, 0, 7)
	fmt.Println("Number of draws:", numOfDraws)
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
			float32(resultNums[n].Normal)/float32(numOfDraws)*100,
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
