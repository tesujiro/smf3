package db

import (
	"encoding/json"
	"log"

	"github.com/garyburd/redigo/redis"
)

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
	//fmt.Printf("func_args:%#v\n", func_args)
	_, err := c.Do("SET", func_args...)
	//fmt.Printf("%s\n", ret)
	return err
}

func db_get(c redis.Conn, key, id string, args ...interface{}) (string, error) {
	func_args := append([]interface{}{key, id}, args...)
	ret, err := c.Do("GET", func_args)
	if err != nil {
		return "", err
	}
	return string(ret.([]byte)), err
}

//func db_retrieve(c redis.Conn, command, key string, args ...interface{}) (string, error) {
func db_retrieve(c redis.Conn, command, key string, args ...interface{}) ([]interface{}, error) {
	func_args := append([]interface{}{key}, args...)
	ret, err := c.Do(command, func_args...)
	if err != nil {
		return nil, err
	}

	records := ret.([]interface{})[1].([]interface{})
	jsonArray := make([]interface{}, len(records))
	for i, b := range records {
		jsonByteArray := b.([]interface{})[1].([]byte)
		var loc interface{}
		err := json.Unmarshal(jsonByteArray, &loc)
		if err != nil {
			return nil, err
		}
		jsonArray[i] = loc
	}

	return jsonArray, err
	/*
		json, err := json.Marshal(jsons)
		if err != nil {
			return "", err
		}
		return string(json), err
	*/
}

func db_scan(c redis.Conn, key string, args ...interface{}) ([]interface{}, error) {
	return db_retrieve(c, "SCAN", key, args...)
}

func db_within(c redis.Conn, key string, s, w, n, e float64, args ...interface{}) ([]interface{}, error) {
	func_args := append([]interface{}{"BOUNDS", s, w, n, e}, args...)
	return db_retrieve(c, "WITHIN", key, func_args...)
}

func db_drop(c redis.Conn, key string) error {
	_, err := c.Do("DROP", key)
	//fmt.Printf("%s\n", ret)
	return err
}
