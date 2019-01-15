package db

import (
	"encoding/json"
	"fmt"
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

func db_set_json(c redis.Conn, key, id, json string) error {
	_, err := c.Do("SET", key, id, "OBJECT", json)
	//fmt.Printf("%s\n", ret)
	return err
}

func db_get(c redis.Conn, key, id string) (string, error) {
	ret, err := c.Do("GET", key, id)
	if err != nil {
		return "", err
	}
	return string(ret.([]byte)), err
}

func db_scan(c redis.Conn, key string) (string, error) {
	ret, err := c.Do("SCAN", key)
	if err != nil {
		return "", err
	}

	records := ret.([]interface{})[1].([]interface{})
	jsons := make([]interface{}, len(records))
	for i, b := range records {
		jsonByteArray := b.([]interface{})[1].([]byte)
		var loc interface{}
		err := json.Unmarshal(jsonByteArray, &loc)
		if err != nil {
			return "", err
		}
		jsons[i] = loc
	}

	json, err := json.Marshal(jsons)
	if err != nil {
		return "", err
	}
	return string(json), err
	/*
		values, err := c.Do("SCAN", key)
		values, err := redis.Values(c.Do("SCAN", key))
		if err != nil {
			return "", err
		}

		fmt.Printf("len(values):%v\n", len(values))
		//for i, val := range values {
		for i, val := range values[1].([]interface{}) {
			fmt.Printf("[%v]: %#v\n", i, val)
		}
	*/
}

func db_drop(c redis.Conn, key string) error {
	ret, err := c.Do("DROP", key)
	fmt.Printf("%s\n", ret)
	return err
}
