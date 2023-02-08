package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-zookeeper/zk"
)

var (
	leaderIsMe bool
	// zkPaths
	leaderCandidaturePath   = "/leader-election-demo/leader-election/candidature"
	leaderParentNodePath    = "/leader-election-demo/leader-election"
	assignedCandidaturePath string
)

const (
	sleepFor = 5
)

func BecomeLeaderThread(ctx context.Context, mch <-chan zk.Event) {
	for {
		err := becomeLeader(ctx, mch)
		if err != nil {
			fmt.Printf("error in becomeLeader - %s\n", err)
			time.Sleep(1 * time.Second)
			continue
		}

		go BeALeader(ctx, mch)

		// wait until session expires or is disconnected
		for {
			mchEvent := <-mch
			if mchEvent.Type == zk.EventSession && (mchEvent.State == zk.StateDisconnected || mchEvent.State == zk.StateExpired) {
				break
			}
		}

		// wait for some time before trying to become leader again
		time.Sleep(1 * time.Second)
	}
}

// becomeLeader returns nil if this node becomes leader
// else returns error.
// TODO: should be invoked whenever session reestabilishes
func becomeLeader(ctx context.Context, mch <-chan zk.Event) error {
	logger := GetLogger(ctx)
	// if assignedCandidaturePath is not empty, then we have already created a node
	// and we are waiting for the leader to be elected
	assignedCandidaturePathExists := false
	if assignedCandidaturePath != "" {
		// check if the assignedCandidaturePath still exists
		exists, _, err := conn.Exists(assignedCandidaturePath)
		if err != nil {
			return fmt.Errorf("error in conn.Exists - %s", err)
		}

		assignedCandidaturePathExists = exists
	}

	if !assignedCandidaturePathExists {
		var err error
		assignedCandidaturePath, err = conn.Create(leaderCandidaturePath, []byte(""), zk.FlagEphemeral|zk.FlagSequence, zk.WorldACL(zk.PermAll))
		if err != nil {
			return fmt.Errorf("error in conn.Create - %s", err)
		}
	}

	logger.Debugf("assignedCandidaturePath - %s", assignedCandidaturePath)

	// for loop breaks when this node becomes leader
	for {
	getChildren:
		allCandidaturePaths, _, err := conn.Children(leaderParentNodePath)
		logger.Debugf("allCandidaturePaths - %+v", allCandidaturePaths)
		if err != nil {
			return fmt.Errorf("error in conn.Children - %s", err)
		}

		// check if im the first child
		if len(allCandidaturePaths) > 0 && (fmt.Sprintf("%s/%s", leaderParentNodePath, allCandidaturePaths[0]) == assignedCandidaturePath) {
			leaderIsMe = true
			fmt.Println("I am the new leader. Aha!")
			break
		}

		prevNodePath := ""
		for _, p := range allCandidaturePaths {
			pFull := fmt.Sprintf("%s/%s", leaderParentNodePath, p)
			if pFull == assignedCandidaturePath {
				break
			}
			prevNodePath = pFull
		}

		exists, _, prevNodeWatcher, err := conn.ExistsW(prevNodePath)
		if err != nil {
			return fmt.Errorf("error in conn.ExistW - %s", err)
		}

		if !exists {
			goto getChildren
		}

		ev := <-prevNodeWatcher
		fmt.Printf("event received from prevNodeWatcher - %+v\n", ev)

		goto getChildren
	}

	return nil
}
