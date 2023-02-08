package main

import "sync"

type EntryLog struct {
	// entries stores all the entries of the distributed log storage system
	entries   []string
	entryLock sync.Mutex
}

var (
	el EntryLog
)

func InitEntryLog() {
	el.entries = []string{}
}

func AddToEntryLog(entry string) error {
	// TODO: zk session invalid check
	el.entryLock.Lock()
	el.entries = append(el.entries, entry)
	el.entryLock.Unlock()
	return nil
}

func GetAllEntryLog() []string {
	entriesCopy := []string{}
	el.entryLock.Lock()
	entriesCopy = append(entriesCopy, el.entries...)
	el.entryLock.Unlock()
	return entriesCopy
}
