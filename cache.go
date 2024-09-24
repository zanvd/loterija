package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type CacheData struct {
	DrawsNum   int                `json:"drawsNum"`
	LastVisit  time.Time          `json:"lastVisit"`
	ResultNums map[int]*resultNum `json:"resultNums"`
}

func Read() (CacheData, error) {
	data := CacheData{
		ResultNums: make(map[int]*resultNum),
	}

	f, err := os.OpenFile("./tmp/cache.json", os.O_RDONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return data, nil
		}
		return data, fmt.Errorf("failed to open cache file: %w", err)
	}
	defer closeFile(f)

	dec := json.NewDecoder(f)
	if err := dec.Decode(&data); err != nil {
		return data, fmt.Errorf("failed to decode cache file: %w", err)
	}

	return data, nil
}

func Write(drawsNum int, resNums map[int]*resultNum) error {
	if err := os.MkdirAll("./tmp", 0744); err != nil {
		return fmt.Errorf("failed to create tmp directory: %w", err)
	}

	f, err := os.OpenFile("./tmp/cache.json", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open cache file: %w", err)
	}
	defer closeFile(f)

	data := CacheData{
		DrawsNum:   drawsNum,
		LastVisit:  time.Now().UTC().Truncate(24 * time.Hour),
		ResultNums: resNums,
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", " ")
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	return nil
}

func closeFile(f *os.File) {
	if err := f.Close(); err != nil {
		fmt.Println("failed to close cache file:", err)
	}
}
