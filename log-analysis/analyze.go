package main

import (
	"fmt"
	"time"
)

const (
	proposalMessage          = "PROPOSING NEW BLOCK ------------------------------------------------"
	twoThirdsMessage         = "[OnCommit] 2/3 Enough commits received"
	oneHundredPercentMessage = "[OnCommit] 100% Enough commits received"
	gracePeriodEndMessage    = "[OnCommit] Commit Grace Period Ended"
	consensusReachedMessage  = "HOORAY!!!!!!! CONSENSUS REACHED!!!!!!!"
)

var blocks = make([]BlockConsensus, 1)
var currentBlock = 0

type BlockConsensus struct {
	proposing          time.Time
	twoThirdsCommitted *time.Time
	allCommitted       *time.Time
	gracePeriodEnd     *time.Time
	consensusReached   *time.Time
}

func analyzeOutput(logMap map[string]interface{}) {
	switch logMap["message"] {
	case proposalMessage:
		//assumes new block will always have logMap["BlockNum"] == current block+1
		time, _ := time.Parse(time.RFC3339, logMap["time"].(string))
		newBlock := BlockConsensus{time, nil, nil, nil, nil}
		blocks = append(blocks, newBlock)

	case twoThirdsMessage:
		time, _ := time.Parse(time.RFC3339, logMap["time"].(string))
		if blocks[currentBlock].twoThirdsCommitted == nil {
			blocks[currentBlock].twoThirdsCommitted = &time
		} else {
			//I don't think the protocol allows us to be multiple blocks behind when we get to this point in logs. Uncertain though.
			blocks[currentBlock+1].twoThirdsCommitted = &time
		}

	case oneHundredPercentMessage:
		time, _ := time.Parse(time.RFC3339, logMap["time"].(string))
		if blocks[currentBlock].allCommitted == nil {
			blocks[currentBlock].allCommitted = &time
		} else {
			blocks[currentBlock+1].twoThirdsCommitted = &time
		}

	case gracePeriodEndMessage:
		time, _ := time.Parse(time.RFC3339, logMap["time"].(string))
		if blocks[currentBlock].gracePeriodEnd == nil {
			blocks[currentBlock].gracePeriodEnd = &time
		} else {
			blocks[currentBlock+1].gracePeriodEnd = &time
		}

	case consensusReachedMessage:
		time, _ := time.Parse(time.RFC3339, logMap["time"].(string))
		blocks[currentBlock].consensusReached = &time
		submitBlockData(blocks[currentBlock])
		currentBlock++
	}
}

func submitBlockData(blockTime BlockConsensus) {
	fmt.Println(blockTime.twoThirdsCommitted)
}
