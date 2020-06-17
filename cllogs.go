package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

var dataform []string = []string{
	time.ANSIC, time.UnixDate,
	time.RubyDate, time.RFC822,
	time.RFC822Z, time.RFC850,
	time.RFC1123, time.RFC1123Z,
	time.RFC3339, time.RFC3339Nano,
	time.Kitchen, time.Stamp,
	time.StampMilli, time.StampMicro,
	time.StampNano, "2006-Jan-02",
	"2006-01-02",
}

var counterlogs int

var counterdeletedlogs int

//ищем временной отрезок в строке по модели первого и последнего числа, встречающегося в строке
func findtimeinstr(str string) (time.Time, error) {
	var result time.Time
	firstindex, secondindex := -1, -1
	for i := 0; i < len(str); i++ {
		_, err := strconv.Atoi(string(str[i]))
		if err != nil {
			continue
		}
		if firstindex == -1 {
			firstindex = i
		}
		secondindex = i
	}
	if firstindex == -1 {
		return result, notfounderror
	}
	result, err := parsetime(str[firstindex : secondindex+1])
	if err != nil {
		return result, err
	}
	return result, nil
}

var notfounderror = errors.New("time is not found here")

func parsetime(str string) (time.Time, error) {
	var result time.Time
	for _, v := range dataform {
		t, err := time.Parse(v, str)
		if err == nil {
			return t, nil
		}
	}
	return result, notfounderror
}

func checklog(filename string, timeForDelete time.Time, path string) {
	ftime, err := findtimeinstr(filename)
	if err != nil {
		return
	}
	mutex.Lock()
	counterlogs++
	mutex.Unlock()
	if ftime.Before(timeForDelete) {
		err := deletefile(path + "/" + filename)
		str := path + "/" + filename
		if err != nil {
			fmt.Printf("%s is not found\n", str)
		}
		mutex.Lock()
		counterdeletedlogs++
		mutex.Unlock()
	}
	wg.Done()
}

var wg sync.WaitGroup
var mutex = &sync.Mutex{}

func main() {
	start := time.Now()

	//	path := "D:/go/src/github.com/puptup/LogCleaner/logs"
	timeF := flag.Int("t", 7, "how many days before deleting the logs")
	pathF := flag.String("p", ".", "path to logs")
	//formats := flag.String("f", "*", "formats to be deleted(all by default)")
	flag.Parse()

	day, _ := time.ParseDuration(strconv.Itoa(*timeF*24) + "h")
	timeForDelete := start.Add(-day)

	files := readDir(*pathF)
	wg.Add(len(files))

	for _, v := range files {
		go checklog(v, timeForDelete, *pathF)
	}
	wg.Wait()
	end := time.Now()
	difference := end.Sub(start)

	fmt.Printf("%d logs were found. Deleted %d. Time spent: %v\n", counterlogs, counterdeletedlogs, difference)
}

var FileIsNotFound = errors.New("file is not found here")

func deletefile(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return FileIsNotFound
	}
	return nil
}

func readDir(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("failed opening directory: %s", err)
	}
	defer file.Close()

	list, _ := file.Readdirnames(0)
	return list
}
