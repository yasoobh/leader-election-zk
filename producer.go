package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-zookeeper/zk"
)

const (
	secondsBetweenTokens = 1
)

func produceTokens(ch chan<- string, mch <-chan zk.Event) {
	iter := 0
	for {
		select {
		case mchEvent := <-mch:
			if mchEvent.Type == zk.EventSession && (mchEvent.State == zk.StateDisconnected || mchEvent.State == zk.StateExpired) {
				return
			}
		default:
			ch <- "I am " + myID + " - the leader. Token no #" + strconv.Itoa(iter)
			iter++
			time.Sleep(time.Second * secondsBetweenTokens)
		}
	}
}

func BeALeader(ctx context.Context, mch <-chan zk.Event) {

	// create a channel to send tokens to
	ch := make(chan string)

	// start producing tokens
	go produceTokens(ch, mch)

	for {
		select {
		// TODO: this won't work. Need to use context.Done instead
		case mchEvent := <-mch:
			if mchEvent.Type == zk.EventSession && (mchEvent.State == zk.StateDisconnected || mchEvent.State == zk.StateExpired) {
				return
			}
		case token := <-ch:
			err := handleToken(ctx, token)
			if err != nil {
				fmt.Printf("error in handleToken - %s\n", err)
			}
		}
	}
}

func handleToken(ctx context.Context, token string) error {
	// store token in entries
	err := AddToEntryLog(token)
	if err != nil {
		return fmt.Errorf("error adding token to entry log - %v, token - %s", err, token)
	}

	// send token to other replicas
	err = SendToReplicas(ctx, token)
	if err != nil {
		return fmt.Errorf("error sending token to replicas - %v, token - %s", err, token)
	}
	fmt.Println("I am the leader - sending token to consumer - ", token)
	return nil
}
