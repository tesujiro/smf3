//package ramble
package main

import (
	"context"
	"fmt"
	"time"

	. "github.com/qedus/osmpbf"
	"github.com/tesujiro/smf3/data/db"
)

type rambler struct {
	db.Location
	curWayID     int64
	prevWayID    int64
	curNodeID    int64
	prevNodeID   int64
	backward     bool // direction
	stop         bool
	speed        float64 // meter per second
	ramble_ratio float64
}

func newRambler(ID int64, lat, lon float64, nodeID, wayID int64) *rambler {
	rambler := rambler{
		Location:     db.Location{ID, lat, lon, time.Now().Format(time.RFC3339)},
		curNodeID:    nodeID,
		prevNodeID:   nodeID,
		curWayID:     wayID,
		prevWayID:    wayID,
		backward:     false,
		stop:         false,
		speed:        0.5,
		ramble_ratio: 0.3,
	}

	if err := (&rambler.Location).Set(); err != nil {
		fmt.Printf("Set Rambler Location error: %v\n", err)
	}
	return &rambler
}

func (r *rambler) walk(ctx context.Context, nodes map[int64]Node, ways map[int64]Way) {
	tick := time.NewTicker(time.Millisecond * time.Duration(500)).C
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-tick:
			way := ways[r.curWayID]
			//node := nodes[r.curNodeID]
			var nextTo func([]int64, int64, bool) int64
			nextTo = func(ids []int64, nid int64, reached bool) int64 {
				//fmt.Printf("ids=%v\n", ids)
				if len(ids) == 0 {
					return 0
				}
				if ids[0] == nid {
					return nextTo(ids[1:], nid, true)
				}
				if _, ok := nodes[ids[0]]; ok && reached {
					return ids[0]
				} else {
					return nextTo(ids[1:], nid, reached)
				}
			}
			randomNodeID := func() int64 {
				for _, node := range nodes {
					return node.ID
				}
				return 0
			}
			var nextNodeID int64
			if nextNodeID = nextTo(way.NodeIDs, r.curNodeID, false); nextNodeID == 0 {
				//fmt.Printf("stop point: --> random \n")
				nextNodeID = randomNodeID()
			} else {
				//fmt.Printf("current way: %#v\n", way.Tags)
				//fmt.Printf("current way node ids: %#v\n", way.NodeIDs)
				//fmt.Printf("current node: %#v \tnext node: %#v\n", r.curNodeID, nextNodeID)
				//fmt.Printf(" %#v \t-> %#v\n", r.curNodeID, nextNodeID)
			}
			r.prevNodeID = r.curNodeID
			r.curNodeID = nextNodeID
			r.Lat = nodes[nextNodeID].Lat
			r.Lon = nodes[nextNodeID].Lon

			if err := (r.Location).Set(); err != nil {
				fmt.Printf("Set Rambler(ID:%v) Location error: %v\n", r.ID, err)
			}
			//fmt.Printf("rambler(ID:%v) tick\n", r.ID)
		}
	}
}
