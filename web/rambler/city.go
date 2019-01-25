//package ramble
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	. "github.com/qedus/osmpbf"
	"github.com/tesujiro/smf3/data/db"
	"github.com/tesujiro/smf3/debug"
)

var number_of_ramblers int

const rampup_msec = 1

type VirtualCity struct {
	nodes      map[int64]Node
	ways       map[int64]Way
	node2way   map[int64][]int64
	lat_center float64
	lon_center float64
	lat_width  float64
	lon_width  float64
	ramblers   []*rambler
	cancelFunc func()
}

const jsonPathWays = "./data/osm/ways.json"
const jsonPathNodes = "./data/osm/nodes.json"

func NewVirtualCity(ctx context.Context, latc, lonc, latw, lonw float64) *VirtualCity {
	var nodes []Node
	var ways []Way

	// Import nodes.json
	data, err := ioutil.ReadFile(jsonPathNodes)
	if err != nil {
		fmt.Printf("Node File Read Error: %v\n", err)
		return nil
	}
	err = json.Unmarshal(data, &nodes)
	if err != nil {
		fmt.Printf("Node File Unmarshal Error: %v\n", err)
		return nil
	}

	// Import ways.json
	data, err = ioutil.ReadFile(jsonPathWays)
	if err != nil {
		fmt.Printf("Way File Read Error: %v\n", err)
		return nil
	}
	err = json.Unmarshal(data, &ways)
	if err != nil {
		fmt.Printf("Way Node File Unmarshal Error: %v\n", err)
		return nil
	}

	// nodesMap
	nodesMap := make(map[int64]Node)
	for _, node := range nodes {
		nodesMap[node.ID] = node
	}
	// waysMap
	waysMap := make(map[int64]Way)
	for _, way := range ways {
		waysMap[way.ID] = way
	}

	// Make map[Node.ID]Way.ID
	node2way := make(map[int64][]int64)
	for _, way := range ways {
		for _, nodeID := range way.NodeIDs {
			if _, ok := nodesMap[nodeID]; ok {
				node2way[nodeID] = append(node2way[nodeID], way.ID)
			}
		}
	}

	// delete invalid node
	for nodeID, _ := range nodesMap {
		if _, ok := node2way[nodeID]; !ok {
			delete(nodesMap, nodeID)
		}
	}

	fmt.Printf("Nodes:%v\n", len(nodesMap))
	fmt.Printf("Ways:%v\n", len(waysMap))
	fmt.Printf("node2way:%v\n", len(node2way))

	vc := VirtualCity{
		nodes:      nodesMap,
		ways:       waysMap,
		node2way:   node2way,
		lat_center: latc,
		lon_center: lonc,
		lat_width:  latw,
		lon_width:  lonw,
	}

	// DROP LOCATION
	if err := db.DropLocation(); err != nil {
		return nil
	}

	for i := 0; i < number_of_ramblers; i++ {
		r := vc.addRambler(db.NewLocationID())
		//fmt.Printf("i=%v lat=%v lon=%v NodeID=%v\tWayID=%v\ttags=%#v\n", i, r.Lat, r.Lon, r.curNodeID, r.curWayID, vc.ways[r.curWayID].Tags)
		go r.walk(ctx, vc.nodes, vc.ways, vc.node2way)
		time.Sleep(rampup_msec * time.Millisecond)
	}

	return &vc
}

func (vc *VirtualCity) addRambler(ID int64) *rambler {
	randomNodeID := func() int64 {
		for _, node := range vc.nodes {
			return node.ID
		}
		return 0
	}

	nodeID := randomNodeID()
	wayID := vc.node2way[nodeID][0]

	rambler := newRambler(ID, vc.nodes[nodeID].Lat, vc.nodes[nodeID].Lon, nodeID, wayID)
	vc.ramblers = append(vc.ramblers, rambler)

	return rambler
}

func (vc *VirtualCity) loop(ctx context.Context, cancel func()) {
	defer cancel()
	//tick := time.NewTicker(time.Millisecond * time.Duration(20)).C
	tick := time.NewTicker(time.Millisecond * time.Duration(500)).C
	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
mainloop:
	for {
		select {
		case <-ctx.Done():
			break mainloop
		case <-tick:
			fmt.Printf("tick\n")
		case s := <-signal_chan:
			fmt.Printf("signal:%v\n", s)
			break mainloop
		}
	}

}

func (vc *VirtualCity) Run(ctx context.Context) error {
	var _cancelOnce sync.Once
	var _cancel func()
	ctx, _cancel = context.WithCancel(ctx)
	cancel := func() {
		_cancelOnce.Do(func() {
			fmt.Printf("Rambler cancel called.\n")
			_cancel()
		})
	}
	vc.cancelFunc = cancel

	go vc.loop(ctx, cancel)

	<-ctx.Done()

	return nil
}

func main() {
	flag.IntVar(&number_of_ramblers, "n", 10, "number of rambler")
	flag.Parse()

	debug.On()
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintf(os.Stderr, "Error:\n%s", err)
			os.Exit(1)
		}
	}()
	//db.ScanLocation() // for testing
	os.Exit(_main())
}

func _main() int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vc := NewVirtualCity(ctx, 0, 0, 0, 0)

	if err := vc.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}
	return 0
}
