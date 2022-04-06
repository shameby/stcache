package cache

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/hashicorp/raft"
)

var cnt uint64 = 0

func GetHandler(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")
	fmt.Fprint(w, GetSlot().Get(key))
}

func SetHandler(w http.ResponseWriter, r *http.Request) {
	if !checkWritePermission {
		fmt.Fprint(w, "follower cant set")
		return
	}
	key := r.FormValue("key")
	val := r.FormValue("val")
	applyFuture := GetRaft().Apply([]byte(key+":"+val), 5 * time.Second)
	if err := applyFuture.Error(); err != nil {
		fmt.Fprint(w, err.Error())
	}
	fmt.Fprint(w, "ok")
}

func StartLoadTest(w http.ResponseWriter, _ *http.Request) {
	if !checkWritePermission {
		fmt.Fprint(w, "follower cant start load test")
		return
	}
	go func() {
		var last uint64 = 0
		for {
			time.Sleep(1*time.Second)
			curr := atomic.LoadUint64(&cnt)
			fmt.Println(curr - last)
			last = curr
		}
	}()
	for i := 0; i < 1000; i++ {
		go func() {
			for {
				n := rand.Intn(100000)
				idx := strconv.Itoa(n)
				key := "ww"+ idx
				val := idx
				applyFuture := GetRaft().Apply([]byte(key+":"+val), 5 * time.Second)
				if err := applyFuture.Error(); err != nil {
					panic(err)
				}
				atomic.AddUint64(&cnt, 1)
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}

	fmt.Fprint(w, "ok")
}

func JoinHandler(w http.ResponseWriter, r *http.Request) {
	addr := r.FormValue("addr")
	id := r.FormValue("id")
	if addr == "" {
		fmt.Fprint(w, "invalid addr")
		return
	}
	addPeerFuture := GetRaft().AddVoter(raft.ServerID(id), raft.ServerAddress(addr), 0, 0)
	if err := addPeerFuture.Error(); err != nil {
		fmt.Fprint(w, "add voter err")
		return
	}
	fmt.Fprint(w, "ok")
	return
}