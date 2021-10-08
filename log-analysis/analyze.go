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

var blocks []BlockConsensus
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
		time, _ := time.Parse(time.RFC3339, logMap["time"].(string))
		newBlock := BlockConsensus{time, nil, nil, nil, nil}
		blocks = append(blocks, newBlock)

	case twoThirdsMessage:
		blockNum := int(logMap["MsgBlockNum"].(float64)) - 1
		time, _ := time.Parse(time.RFC3339, logMap["time"].(string))
		blocks[blockNum].twoThirdsCommitted = &time

	case oneHundredPercentMessage:
		time, _ := time.Parse(time.RFC3339, logMap["time"].(string))
		if blocks[currentBlock].allCommitted == nil {
			blocks[currentBlock].allCommitted = &time
		} else {
			blocks[currentBlock+1].twoThirdsCommitted = &time
		}

	case gracePeriodEndMessage:
		time, _ := time.Parse(time.RFC3339, logMap["time"].(string))
		blocks[currentBlock].gracePeriodEnd = &time
		if blocks[currentBlock].consensusReached != nil {
			go submitBlockData(blocks[currentBlock], currentBlock)
			currentBlock++
		}

	case consensusReachedMessage:
		blockNum := int(logMap["blockNum"].(float64)) - 1
		time, _ := time.Parse(time.RFC3339, logMap["time"].(string))
		blocks[blockNum].consensusReached = &time
		if blocks[currentBlock].gracePeriodEnd != nil {
			submitBlockData(blocks[currentBlock], currentBlock)
			currentBlock++
		}
	}
}

func submitBlockData(blockTime BlockConsensus, block int) {
	proposingTime := blocks[block].proposing
	/* 	twoThirdsTime := blocks[block].twoThirdsCommitted
	   	fmt.Printf("Time between Proposal and 2/3 Commits for block %v was %v\n", block+1, twoThirdsTime.Sub(proposingTime))

	   		if blocks[block].allCommitted != nil {
	   		allCommittedTime := blocks[block].allCommitted
	   		fmt.Printf("Time between 2/3 Commits and 100%% Commits for block %v was %v\n", block+1, allCommittedTime.Sub(*twoThirdsTime))
	   	}
	*/
	consensusReachedTime := blocks[block].consensusReached
	//	fmt.Printf("Time between 2/3 Commits and Consensus reached for block %v was %v\n", block+1, consensusReachedTime.Sub(*twoThirdsTime))
	fmt.Printf("Time between Proposal and Consensus Reached for block %v was %v\n", block+1, consensusReachedTime.Sub(proposingTime))

	/* 	gracePeriodTime := blocks[block].gracePeriodEnd
	   	fmt.Printf("Time after consensus reached and grace period ending for block %v was %v\n", block+1, gracePeriodTime.Sub(*consensusReachedTime))
	*/
}
