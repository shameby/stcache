package cache

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/raft"
)

var checkWritePermission bool

var raftC *raft.Raft

func GetRaft() *raft.Raft {
	return raftC
}

type Options struct {
	ServerId  string
	TcpAddr   string
	JoinAddr  string
	Bootstrap bool
}

func Init(opt *Options) {
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(opt.ServerId)
	raftConfig.SnapshotInterval = 20 * time.Second
	raftConfig.SnapshotThreshold = 2
	leaderNotifyCh := make(chan bool, 1)
	raftConfig.NotifyCh = leaderNotifyCh
	go func() {
		for range leaderNotifyCh {
			checkWritePermission = true
		}
	}()

	// logStore, err := boltdb.NewBoltStore(fmt.Sprintf("./bolt/raft-log%s.bolt", opt.ServerId))
	logStore, err := NewBadgerStore(fmt.Sprintf("./badger/raft-log%s.bolt", opt.ServerId))
	if err != nil {
		return
	}
	stableStore, err := NewBadgerStore(fmt.Sprintf("./badger/stable-log%s.bolt", opt.ServerId))
	if err != nil {
		return
	}
	//stableStore, err := boltdb.NewBoltStore(fmt.Sprintf("./bolt/stable-log%s.bolt", opt.ServerId))

	// snapshotStore := raft.NewInmemSnapshotStore()
	snapshotStore, err := raft.NewFileSnapshotStore(fmt.Sprintf("./ss%s", opt.ServerId), 1, os.Stderr)
	if err != nil {
		panic(err)
	}

	transport, err := newRaftTransport(opt)
	if err != nil {
		panic(err)
	}

	raftC, err = raft.NewRaft(raftConfig, &FSM{}, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		panic(err)
	}

	if opt.Bootstrap {
		cfg := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raftConfig.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		f := raftC.BootstrapCluster(cfg)
		if err := f.Error(); err != nil {
			panic(err)
		}
	}

	if opt.JoinAddr != "" {
		if err := joinRaftCluster(opt); err != nil {
			panic(err)
		}
	}
}

func newRaftTransport(opt *Options) (*raft.NetworkTransport, error) {
	address, err := net.ResolveTCPAddr("tcp", opt.TcpAddr)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(address.String(), address, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}
	return transport, nil
}

func joinRaftCluster(opt *Options) error {
	url := fmt.Sprintf("http://%s/join?addr=%s&id=%s", opt.JoinAddr, opt.TcpAddr, opt.ServerId)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if string(body) != "ok" {
		return errors.New("err join cluster")
	}
	return nil
}
