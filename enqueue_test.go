package workers

import "testing"

func TestEnqueue(t *testing.T) {
	Configure(RedisOptions{
		Addr: "127.0.0.1:6379",
	}, RedisOptions{
		KeyWord: "fanyangyang",
		Addr:    "127.0.0.1:6379",
	})
	if _, err := Enqueue("", "msg", "MsgWorker", []interface{}{map[string]interface{}{"key": "key1", "value": "value1"}}); err != nil {
		t.Error(err)
	}

	if _, err := Enqueue("fanyangyang", "msg", "MsgWorker", []interface{}{map[string]interface{}{"key": "key2", "value": "value2"}}); err != nil {
		t.Error(err)
	}
}
