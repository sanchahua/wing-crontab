package service

import (
	"google.golang.org/grpc/metadata"
	"log"
	"os"
)

func retrieveFromMeta(md metadata.MD, key string) string {
	if vals, ok := md[key]; ok && len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func checkEnv(key string) string {

	v := os.Getenv(key)
	log.Printf("[xlsoa] Checking env '%v'=>'%v'\n", key, v)

	return v
}
