package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

const startDate = "2018-03-24"
const urlTemplate = "https://www.easyjet.com/ejcms/nocache/api/lowestfares/get/?originIata=%s&destinationIata=%s&displayCurrencyId=4&languageCode=en&startDate=%s"

const defaultRootDir = "/flightdata"
const responseDir = "/responses"
const pollDir = "/polls"

func makeRequest(url string) ([]byte, error) {
	fmt.Printf("Making request to: [%s]\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}

func makeURL(origin string, dest string, date string) string {
	return fmt.Sprintf(urlTemplate, origin, dest, date)
}

func computeChecksum(content []byte) string {
	hashtype := sha1.New()
	hashtype.Write(content)
	hashInBytes := hashtype.Sum(nil)[:20]
	return hex.EncodeToString(hashInBytes)
}

func processRoute(responseDirPath string, pollDirPath string, origin string, dest string, date string) error {
	respBody, err := makeRequest(makeURL(origin, dest, date))
	if err != nil {
		return err
	}
	checksum := computeChecksum(respBody)
	datedPollDirPath := pollDirPath + "/" + time.Now().Format("2006-0102")
	os.MkdirAll(datedPollDirPath, os.ModePerm)
	filename := fmt.Sprintf(datedPollDirPath+"/%s-%s-%s-%s", origin, dest, time.Now().Format("2006-01-02-150405"), checksum)
	ioutil.WriteFile(responseDirPath+"/"+checksum, respBody, 0644)
	ioutil.WriteFile(filename, []byte(checksum+"\n"), 0644)
	return nil
}

func main() {

	var rootDir = defaultRootDir
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}
	responseDirPath := rootDir + responseDir
	pollDirPath := rootDir + pollDir

	fmt.Println("Responses out to: " + responseDirPath)
	fmt.Println("Polls out to: " + pollDirPath)
	os.MkdirAll(responseDirPath, os.ModePerm)
	os.MkdirAll(pollDirPath, os.ModePerm)

	cities := [...]string{
		"ALC",
		"BCN",
		"BJV",
		"BOD",
		"CFU",
		"HER",
		"DLM",
		"FAO",
		"LPA",
		"EFL",
		"ACE",
		"AGP",
		"MAH",
		"MJV",
		"NAP",
		"NCE",
		"PUY",
		"KEF",
		"OLB",
		"CTA",
		"TFS",
		"TLS",
		"ZTH"}
	var wg sync.WaitGroup
	wg.Add(len(cities))
	for _, city := range cities {
		go func(city string) {
			defer wg.Done()
			processRoute(responseDirPath, pollDirPath, "BRS", city, startDate)
			processRoute(responseDirPath, pollDirPath, city, "BRS", startDate)
		}(city)
	}
	wg.Wait()
}
