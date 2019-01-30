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
	ret, err := c.Do(command, func_args...)
	if err != nil {
		return nil, err
	}

	features := []GeoJsonFeature{}
	// ret
	// ret.([]interface{})[0] : cursor number
	// ret.([]interface{})[1] : objects
	objects := ret.([]interface{})[1].([]interface{}) //objects
	if len(objects) > 0 {
		features = make([]GeoJsonFeature, len(objects))
		for i, object := range objects {
			// object
			// object.([]interface{})[0]: id ([]byte)
			// object.([]interface{})[1]: json ([]byte)
			// object.([]interface{})[2]: field and value pairs ([]interface{}) ex. [ start 123 end 456 ]
			b := object.([]interface{})[1].([]byte) //json
			//fmt.Printf("b:%#v\n", string(b))
			err = json.Unmarshal(b, &features[i])
			if err != nil {
				return nil, err
			}
		}
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
	//fmt.Printf("%s\n", ret)
	return err
}
