//package ramble
package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	. "github.com/qedus/osmpbf"
	"github.com/tesujiro/smf3/data/db"
	"github.com/tesujiro/smf3/debug"
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
		Location:     db.Location{ID: ID, Lat: lat, Lon: lon, Time: time.Now().Unix()},
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

func (r *rambler) walk(ctx context.Context, nodes map[int64]Node, ways map[int64]Way, node2way map[int64][]int64) {
	tick := time.NewTicker(time.Millisecond * time.Duration(500)).C
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-tick:
			restNodes := func(nodeID int64, wayID int64) int {
				for i, nid := range ways[wayID].NodeIDs {
					if nid == nodeID {
						return len(ways[wayID].NodeIDs) - i - 1
						//return len(ways[wayID].NodeIDs) - i - 2 // avoid short loop
					}
				}
				return 0
			}
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
			anotherWayID := func() int64 {
				if len(node2way[r.curNodeID]) < 2 {
					return r.curWayID
				}
				//debug.Printf("take another way\n")
				others := []int64{}
				for _, wayID := range node2way[r.curNodeID] {
					//if wayID != r.curWayID && wayID != r.prevWayID {
					if restNodes(r.curNodeID, wayID) > 0 {
						others = append(others, wayID)
					}
				}
				switch len(others) {
				case 0:
					debug.Printf("take another way --> failed\n")
					return r.curWayID
				case 1:
					if rand.Intn(10) == 0 { // avoid loop
						return r.curWayID
					} else {
						return others[0]
					}
				default:
					way := others[rand.Intn(len(others))]
					debug.Printf("take way current:%v next:%v ways:%v change:%v\n", r.curWayID, way, len(others), r.curWayID != way)
					return way
				}
				/*
					if len(others) == 0 {
						debug.Printf("take another way --> failed\n")
						return r.curWayID
					} else {
						way := others[rand.Intn(len(others))]
						debug.Printf("take way current:%v next:%v ways:%v change:%v\n", r.curWayID, way, len(others), r.curWayID != way)
						return way
					}
				*/
			}
			randomNodeID := func() int64 {
				for _, node := range nodes {
					return node.ID
				}
				return 0
			}
			uniqNodes := func(nodes []int64) []int64 {
				uniq := make([]int64, 0, len(nodes))
				encountered := map[int64]bool{}
				for i := 0; i < len(nodes); i++ {
					if !encountered[nodes[i]] {
						uniq = append(uniq, nodes[i])
						encountered[nodes[i]] = true
					}
				}
				return uniq
			}
			var nextNodeID int64

			// TODO: if other way exist , change current way.
			nextWayID := anotherWayID()
			r.prevWayID = r.curWayID
			r.curWayID = nextWayID
			if restNodes(r.curNodeID, r.curWayID) > 0 {
				r.backward = false
			} else {
				r.backward = true
				debug.Printf("set backward\n")
			}

			var nodeIDs []int64
			if !r.backward {
				nodeIDs = uniqNodes(ways[r.curWayID].NodeIDs)
			} else {
				reverse := func(a []int64) (opp []int64) {
					for i := len(a)/2 - 1; i >= 0; i-- {
						opp := len(a) - 1 - i
						a[i], a[opp] = a[opp], a[i]
					}
					return opp
				}
				nodeIDs = uniqNodes(reverse(ways[r.curWayID].NodeIDs))
				//_ = reverse(ways[r.curWayID].NodeIDs)
				//nodeIDs = ways[r.curWayID].NodeIDs
			}
			//if nextNodeID = nextTo(nodeIDs, r.curNodeID, false); nextNodeID == 0 || rand.Intn(500) == 0 { // random once in 500 steps
			if nextNodeID = nextTo(nodeIDs, r.curNodeID, false); nextNodeID == 0 { // random once in 500 steps
				debug.Printf("stop point: --> random \n")
				debug.Printf("curNode(ID:%v) curWay(ID:%v):%v\n", r.curNodeID, r.curWayID, ways[r.curWayID].NodeIDs)
				nextNodeID = randomNodeID()
				nextWayID := node2way[nextNodeID][0]
				r.prevWayID = r.curWayID
				r.curWayID = nextWayID
				if restNodes(r.curNodeID, r.curWayID) > 0 {
					r.backward = false
				} else {
					r.backward = true
					//debug.Printf("set backward")
				}
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
