package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-zookeeper/zk"
)

const (
	membershipPath = "/leader-election-demo/membership"
)

type member struct {
	ID   string `json:"id"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

var (
	replicas     []member
	replicasLock sync.Mutex
)

func validateParentExists() error {
	// check if membershipPath exists
	exists, _, err := conn.Exists(membershipPath)
	if err != nil {
		return fmt.Errorf("failed to check if parent znode exists - %s", err)
	}

	if exists {
		return nil
	}

	_, err = conn.Create(membershipPath, []byte{}, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		if err == zk.ErrNodeExists {
			return nil
		}

		return fmt.Errorf("failed to create parent znode - %s", err)
	}

	return nil
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		panic("failed to get hostname - " + err.Error())
	}

	return hostname
}

func getPort() int {
	// returns the port defined in main.go
	return port
}

// becomeMember creates an ephemeral znode at membershipPath
func becomeMember(ctx context.Context) error {
	if err := validateParentExists(); err != nil {
		return err
	}

	mem := member{
		ID:   myID,
		Host: getHostname(),
		Port: port,
	}

	serMem, err := json.Marshal(mem)
	if err != nil {
		return fmt.Errorf("failed to marshal member - %s", err)
	}

	_, err = conn.Create(membershipPath+"/"+myID, []byte(serMem), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		if err == zk.ErrNodeExists {
			log := GetLogger(ctx)
			log.Debugf("membership znode already exists - %s", err)
			return nil
		}
		return fmt.Errorf("failed to create membership znode - %s", err)
	}

	return nil
}

func MembershipThread(ctx context.Context) {
	log := GetLogger(ctx)
	for {
		err := becomeMember(ctx)
		if err != nil {
			log.Errorf("failed to become member - %s", err)
			time.Sleep(1 * time.Second)
			continue
		}

		mems, watcher, err := getMembers(ctx)
		if err != nil {
			log.Errorf("failed to get members - %s", err)
			time.Sleep(1 * time.Second)
			continue
		}

		log.Debugf("updated members - %v", mems)

		replicasLock.Lock()
		replicas = mems
		replicasLock.Unlock()

		for {
			ev := <-watcher
			if ev.Type == zk.EventNodeChildrenChanged {
				log.Debugf("membership changed - %v", ev)
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// getMembers returns a list of members
func getMembers(ctx context.Context) ([]member, <-chan zk.Event, error) {
	mems := make([]member, 0)

	children, _, watcher, err := conn.ChildrenW(membershipPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get children of membership path - %s", err)
	}

	for _, child := range children {
		childPath := fmt.Sprintf("%s/%s", membershipPath, child)

		data, _, err := conn.Get(childPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get data of child - %s", err)
		}

		var mem member
		err = json.Unmarshal(data, &mem)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal member - %s", err)
		}

		mems = append(mems, mem)
	}

	return mems, watcher, nil
}
