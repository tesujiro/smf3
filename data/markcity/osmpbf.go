package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"

	"github.com/qedus/osmpbf"
)

const (
	lat_min = 35.6570
	lat_max = 35.6592
	lon_min = 139.6960
	lon_max = 139.6990
)

//const filepath = "/Users/tesujiro/Downloads/JP"
const filepath = "/Users/tesujiro/Downloads/kanto-latest.osm.pbf"

type location struct {
	lat float64
	lon float64
}

func inArea(lat, lon float64) bool {
	return (lat_min <= lat && lat <= lat_max) &&
		(lon_min <= lon && lon <= lon_max)
}

func getNodes() []*osmpbf.Node {
	var nodes []*osmpbf.Node

	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	d := osmpbf.NewDecoder(f)
	d.SetBufferSize(osmpbf.MaxBlobSize)
	err = d.Start(runtime.GOMAXPROCS(-1))
	if err != nil {
		log.Fatal(err)
	}

	for {
		if v, err := d.Decode(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			switch v := v.(type) {
			case *osmpbf.Node:
				// Process Node v.
				/*
					if len(v.Tags) > 2 {
						//fmt.Printf("Node: %#v\n", v)
						//fmt.Printf("Node: Lat %#v Lon %#v\n", v.Lat, v.Lon)
					}
				*/
				if inArea(v.Lat, v.Lon) {
					nodes = append(nodes, (*osmpbf.Node)(v))
					//fmt.Printf("Node: %#v\n", v)
				} else {
					break
				}
			case *osmpbf.Way:
			case *osmpbf.Relation:
			default:
				log.Fatalf("unknown type %T\n", v)
			}
		}
	}
	return nodes
}

func getWays(nodes []*osmpbf.Node) []*osmpbf.Way {
	var ways []*osmpbf.Way

	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	d := osmpbf.NewDecoder(f)
	d.SetBufferSize(osmpbf.MaxBlobSize)
	err = d.Start(runtime.GOMAXPROCS(-1))
	if err != nil {
		log.Fatal(err)
	}

	node_map := make(map[int64]location, len(nodes))
	for _, n := range nodes {
		node_map[n.ID] = location{lat: n.Lat, lon: n.Lon}
	}

	for {
		if v, err := d.Decode(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			switch v := v.(type) {
			case *osmpbf.Node:
			case *osmpbf.Way:
				w := (*osmpbf.Way)(v)
				for _, id := range w.NodeIDs {
					if _, ok := node_map[id]; ok {
						ways = append(ways, w)
						//fmt.Printf("Way(node:%v): %#v\n", len(w.NodeIDs), w)
						//fmt.Printf("Way(node:%v): %#v\n", len(w.NodeIDs), w.Tags)
						// Dogenzaka
						if w.ID == 32621715 {
							fmt.Printf("Way(node:%v): %#v\n", len(w.NodeIDs), w)
							for _, v := range w.NodeIDs {
								if location, ok := node_map[v]; ok {
									fmt.Printf("\"lat\":%v, \"lon\": %v\n", location.lat, location.lon)
								}
							}
						}
						break
					}
				}
			case *osmpbf.Relation:
			default:
				log.Fatalf("unknown type %T\n", v)
			}
		}
	}
	return ways
}

func main() {

	nodes := getNodes()
	fmt.Printf("Nodes: %v\n", len(nodes))
	ways := getWays(nodes)
	fmt.Printf("Ways: %v\n", len(ways))
}
