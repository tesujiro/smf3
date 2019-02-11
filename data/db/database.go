package db

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

const MAX_NUMBER = 100000000

var pool *redis.Pool

func init() {
	fmt.Printf("Create Pool\n")
	pool = &redis.Pool{
		MaxIdle:     1024,
		MaxActive:   1024,
		IdleTimeout: 20 * time.Second,
		Dial:        db_connect,
		//Wait:        true,
	}
}

func db_connect() (redis.Conn, error) {
	c, err := redis.Dial("tcp", ":9851")
	if err != nil {
		log.Fatalf("Could not connect: %v\n", err)
		return nil, err
	}
	return c, nil
}

func db_set_json(c redis.Conn, key, id, json string, args ...interface{}) error {
	// see Conn.Do function func signature
	func_args := append([]interface{}{key, id}, args...)
	func_args = append(func_args, []interface{}{"OBJECT", json}...)
	//fmt.Printf("db_retrieve: func_args(%#v)\n", func_args)
	_, err := c.Do("SET", func_args...)
	//fmt.Printf("%s\n", ret)
	return err
}

func db_jset(c redis.Conn, key, id, path string, value interface{}, args ...interface{}) error {
	func_args := append([]interface{}{key, id, path, value}, args...)
	//fmt.Printf("func_args(%#v)\n", func_args)
	_, err := c.Do("JSET", func_args...)
	//fmt.Printf("%s\n", ret)
	return err
}

func db_get(c redis.Conn, key, id string) ([]byte, error) {
	ret, err := c.Do("GET", key, id)
	if err != nil {
		return nil, err
	}

	if b, ok := ret.([]byte); !ok {
		return nil, nil
	} else {
		return b, err
	}
}

func db_del(c redis.Conn, key, id string) error {
	_, err := c.Do("DEL", key, id)
	if err != nil {
		return err
	}
	return nil
}

func db_retrieve(c redis.Conn, command, key string, args ...interface{}) ([]GeoJsonFeature, error) {
	func_args := append([]interface{}{key}, args...)
	//fmt.Printf("db_retrieve: func_args(%#v)\n", func_args)
	results, err := redis.Values(c.Do(command, func_args...))
	if err != nil {
		return nil, fmt.Errorf("Could not SCAN: %v", err)
	}

	var cursor int
	var members []interface{}
	_, err = redis.Scan(results, &cursor, &members)
	if err != nil {
		return nil, fmt.Errorf("scan result error: %v", err)
	}

	features := make([]GeoJsonFeature, len(members))
	i := 0
	for len(members) > 0 {
		// pick up one record from results as []interface{}
		var object []interface{}
		members, err = redis.Scan(members, &object)
		if err != nil {
			return nil, fmt.Errorf("scan record error: %v", err)
		}
		// scan columns from one record -> [id,json],fields
		var id []byte
		var geojson []byte
		_, err := redis.Scan(object, &id, &geojson)
		if err != nil {
			return nil, fmt.Errorf("scan columns error: %v", err)
		}

		// unmarshal geojson string to struct
		err = json.Unmarshal(geojson, &features[i])
		if err != nil {
			return nil, fmt.Errorf("unmarshal json error: %v", err)
		}
		i++
	}

	return features, nil
}

func db_scan(c redis.Conn, key string, args ...interface{}) ([]GeoJsonFeature, error) {
	return db_retrieve(c, "SCAN", key, args...)
}

func db_withinBounds(c redis.Conn, key string, s, w, n, e float64, args ...interface{}) ([]GeoJsonFeature, error) {
	func_args := append([]interface{}{"LIMIT", MAX_NUMBER}, args...)
	func_args = append(func_args, []interface{}{"BOUNDS", s, w, n, e}...)
	//fmt.Printf("db_withinBounds: func_args(%#v)\n", func_args)
	return db_retrieve(c, "WITHIN", key, func_args...)
}

func db_withinCircle(c redis.Conn, key string, lat, lon, meters float64, args ...interface{}) ([]GeoJsonFeature, error) {
	func_args := append([]interface{}{"LIMIT", MAX_NUMBER}, args...)
	func_args = append(func_args, []interface{}{"CIRCLE", lat, lon, meters}...)
	//fmt.Printf("db_withinCircle: func_args(%#v)\n", func_args)
	return db_retrieve(c, "WITHIN", key, func_args...)
}

func db_drop(c redis.Conn, key string) error {
	_, err := c.Do("DROP", key)
	return err
}

func db_sethook(c redis.Conn, name, endpoint string, args ...interface{}) error {
	func_args := append([]interface{}{name, endpoint}, args...)
	//fmt.Printf("db_sethook: func_args(%#v)\n", func_args)
	_, err := c.Do("SETHOOK", func_args...)
	return err
}

func db_pdelhook(c redis.Conn, pattern string) error {
	_, err := c.Do("PDELHOOK", pattern)
	return err
}
