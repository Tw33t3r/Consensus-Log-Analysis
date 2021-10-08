package main

import (
	"bufio"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Consensus log analysis error: Missing file input")
		return
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := bufio.NewScanner(file)
	for reader.Scan() {
		parsedOutput := parse(reader.Text())
		analyzeOutput(parsedOutput)
	}

	if err := reader.Err(); err != nil {
		log.Fatal(err)
	}
}