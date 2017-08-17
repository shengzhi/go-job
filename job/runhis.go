package job

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type RunHistory struct {
	StartTime, CompletedTime time.Time
	Elased                   int64
	Result                   string
	ErrMsg                   string
}

func (rh RunHistory) String() string {
	return fmt.Sprintf("%s-%s %ds %s %s",
		rh.StartTime.Format("01/02 15:04:05"),
		rh.CompletedTime.Format("01/02 15:04:05"),
		rh.Elased, rh.Result, rh.ErrMsg)
}

type RunHisList struct {
	list       []RunHistory
	index, cap int
	lock       *sync.Mutex
}

func NewRunHisList(capacity int) *RunHisList {
	return &RunHisList{
		cap:   capacity,
		list:  make([]RunHistory, 0, capacity),
		index: 0,
		lock:  &sync.Mutex{},
	}
}

func (h *RunHisList) add(his RunHistory) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if len(h.list) < h.cap {
		h.list = append(h.list, his)
		return
	}

	h.list[h.index] = his
	h.index++
	if h.index >= h.cap {
		h.index = 0
	}
}

func (h *RunHisList) His() []RunHistory {
	slice := append(make([]RunHistory, 0, h.cap), h.list[:]...)
	sort.Slice(slice, func(i int, j int) bool {
		return slice[i].StartTime.After(slice[j].StartTime)
	})
	return slice
}
