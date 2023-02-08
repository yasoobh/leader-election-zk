package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/google/uuid"
)

var (
	conn *zk.Conn
)

var myID string

func init() {
	myID = uuid.New().String()
}

type zkLogger struct{}

func (zl zkLogger) Printf(format string, a ...interface{}) {
	fmt.Println("zk logger - ", fmt.Sprintf(format, a...))
}

const (
	port = 8080
)

func main() {
	zookeeperURLs := os.Getenv("ZOOKEEPER_URLS")
	servers := strings.Split(zookeeperURLs, ",")

	var log logger
	log = fmtLogger{}

	log.Debugf("connecting to zookeeper at %s", servers)

	c, mch, err := zk.Connect(
		servers,
		time.Duration(5*time.Second),
		zk.WithLogger(zkLogger{}),
	)
	if err != nil {
		panic("zk connect failed with error - " + err.Error())
	}

	conn = c

	ctx := WithLogger(context.Background(), log)

	log.Infof("myID - %s, hostname - %s", myID, getHostname())

	err = ensureZkPathsExist()
	if err != nil {
		panic("failed to ensure zk paths exist - " + err.Error())
	}

	fmt.Println("my id - ", myID)

	InitEntryLog()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("panic recovered for BecomeLeaderThread - ", r)
			}
		}()
		MembershipThread(ctx)
	}()

	// sleep for a second for replicas to register themselves
	time.Sleep(1 * time.Second)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("panic recovered for BecomeLeaderThread - ", r)
			}
		}()
		BecomeLeaderThread(ctx, mch)
	}()

	InitializeReceiver(port)
}

func ensureZkPathsExist() error {
	paths := []string{
		"/leader-election-demo",
		"/leader-election-demo/leader-election",
		"/leader-election-demo/membership",
	}

	for _, path := range paths {
		exists, _, err := conn.Exists(path)
		if err != nil {
			return fmt.Errorf("failed to check if parent znode exists - %s", err)
		}

		if exists {
			continue
		}

		_, err = conn.Create(path, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			if err == zk.ErrNodeExists {
				continue
			}

			return fmt.Errorf("failed to create parent znode - %s", err)
		}
	}

	return nil
}
