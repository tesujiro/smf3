package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"

	"github.com/qedus/osmpbf"
)

const (
	// Shibuya MarkCity
	lat_center = 35.6581
	lon_center = 139.6975
	//lat_width  = 0.0011
	lat_width = 0.0055
	//lat_width = 0.011
	//lon_width  = 0.0015
	lon_width = 0.0060
	//lon_width = 0.015
	lat_min = lat_center - lat_width
	lat_max = lat_center + lat_width
	lon_min = lon_center - lon_width
	lon_max = lon_center + lon_width
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

func isFootway(way *osmpbf.Way) bool {
	if _, ok := way.Tags["building"]; ok {
		return false
	}
	if _, ok := way.Tags["leisure"]; ok {
		return true
	}
	//return true
	return way.Tags["highway"] == "footway" ||
		way.Tags["highway"] == "pedestrian" ||
		way.Tags["highway"] == "steps" ||
		way.Tags["highway"] == "path" ||
		way.Tags["highway"] == "unclassified" ||
		way.Tags["highway"] == "primary" ||
		way.Tags["highway"] == "trunk" ||
		way.Tags["highway"] == "trunk_link" ||
		way.Tags["highway"] == "tertiary" ||
		way.Tags["highway"] == "service" ||
		way.Tags["highway"] == "residential" ||
		way.Tags["sidewalk"] == "both" ||
		way.Tags["sidewalk"] == "left" ||
		way.Tags["sidewalk"] == "right" ||
		way.Tags["foot"] == "yes" ||
		way.Tags["indoor"] == "yes" ||
		way.Tags["bridge"] == "viaduct" ||
		way.Tags["public_transport"] == "platform" ||
		way.Tags["railway"] == "platform"
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
						if !isFootway(w) {
							continue
						}
						ways = append(ways, w)
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

func makeJson(object interface{}, path string) error {
	data, err := json.Marshal(object)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, data, os.ModePerm)
	if err != nil {
		return err
	}
	fmt.Printf("Write file succeeded: %v\n", path)
	return nil
}

type jsonmap map[string]interface{}

func clientInfo(node_map map[int64]*osmpbf.Node, ways []*osmpbf.Way) interface{} {
	info := make([]jsonmap, len(ways)) //TODO: pointer?? like []*jsonmap for performance
	for k, w := range ways {
		jm := jsonmap{}
		jm["ID"] = w.ID
		jm["Tags"] = w.Tags
		nodes := make([]*osmpbf.Node, len(w.NodeIDs))
		for l, id := range w.NodeIDs {
			if v, ok := node_map[id]; ok {
				nodes[l] = v
			} // TODO: nil ??
		}
		jm["Nodes"] = nodes
		info[k] = jm
	}
	return info
}

func main() {

	nodes := getNodes()
	fmt.Printf("Nodes: %v\n", len(nodes))

	ways := getWays(nodes)
	fmt.Printf("Ways: %v\n", len(ways))

	node_map := make(map[int64]*osmpbf.Node, len(nodes))
	for _, n := range nodes {
		node_map[n.ID] = n
	}

	makeJson(clientInfo(node_map, ways), "ways.json")
}
