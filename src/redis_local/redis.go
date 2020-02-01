package redis_local

import (

	// Import the redigo/redis package.
	"github.com/gomodule/redigo/redis"
	"fmt"
)

var redis_conn redis.Conn;

func init() {
	redis_conn, _ = redis.Dial("tcp", "redis:6379")
}
func GetStringKeyFromMap(main_key string, sub_key string) string{
	// Establish a connection to the Redis server listening on port
	// 6379 of the local machine. 6379 is the default port, so unless
	// you've already changed the Redis configuration file this should
	// work.
	// Importantly, use defer to ensure the connection is always
	// properly closed before exiting the main() function.
	

	// Send our command across the connection. The first parameter to
	// Do() is always the name of the Redis command (in this example
	// HMSET), optionally followed by any necessary arguments (in this
	// example the key, followed by the various hash fields and values).
	// _, err = conn.Do("HMSET", "auth_token:1212", "qbol_user", "1", "account_id", "1")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	if redis_conn == nil {
		redis_conn, _ = redis.Dial("tcp", "redis:6379")
		fmt.Println("CREATED")
	} else {
		fmt.Println("REUSED")
	}
	//redis_conn, _ = redis.Dial("tcp", "127.0.0.1:6379")
	val, err := redis.String(redis_conn.Do("HGET", main_key, sub_key))
	if err != nil {
		return "ERROR fetching data"
	}
	return val
}
