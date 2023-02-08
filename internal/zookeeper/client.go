package zookeeper

// // wrapper over zk client implementation
// import (
// 	"time"

// 	"github.com/go-zookeeper/zk"
// )

// type zkClient interface {
// 	// Create creates the znode with data and returns the path, and error if any
// 	Create(path string, data []byte, flags int32) (string, error)
// 	Exists(path string) (bool, error)
// 	ExistsW(path string)
// }

// var (
// 	conn   *zk.Conn
// 	mainCh <-chan zk.Event
// )

// func Init() {
// 	servers := []string{"zoo1:2181,zoo2:2181,zoo3:2181"}
// 	c, mch, err := zk.Connect(servers, time.Duration(60*time.Second), nil)
// 	if err != nil {
// 		panic("zk connect failed with error - " + err.Error())
// 	}

// 	conn = c
// 	mainCh = mch
// }
