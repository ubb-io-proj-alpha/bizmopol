package service

import "log"

func debugLog(method string, kvPairs ...interface{}) {
	args := make([]interface{}, 0, len(kvPairs)+1)
	args = append(args, method)
	args = append(args, kvPairs...)
	log.Println(args...)
}

func debugLogResult(method string, err error, kvPairs ...interface{}) {
	if err != nil {
		args := make([]interface{}, 0, len(kvPairs)+2)
		args = append(args, method, "ERROR:", err)
		args = append(args, kvPairs...)
		log.Println(args...)
		return
	}
	args := make([]interface{}, 0, len(kvPairs)+2)
	args = append(args, method, "OK")
	args = append(args, kvPairs...)
	log.Println(args...)
}
