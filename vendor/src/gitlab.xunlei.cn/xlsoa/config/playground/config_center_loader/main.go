package main

import (
	"flag"
	"fmt"
	"gitlab.xunlei.cn/xlsoa/config"
	"log"
	"os"
)

var (
	help = false
	addr = ""
	name = ""
)

type myconfig struct {
	Name   string
	Server struct {
		Addr string
		Port int
	}
	Mysql struct {
		Host     string
		Port     int
		User     string
		Password string
	}
	Token struct {
		Ttl int32
	}
}

func main() {

	var err error

	flag.BoolVar(&help, "help", false, "Help message")
	flag.StringVar(&addr, "addr", "localhost:8500", "Consul address")
	flag.StringVar(&name, "name", "xlsoa-core-auth", "Service name")
	flag.Parse()
	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if addr == "" || name == "" {
		log.Println("Addr and name requied")
		flag.PrintDefaults()
		os.Exit(0)
	}

	prefix := fmt.Sprintf("config/%v", name)
	var loader config.Loader
	loader = config.NewConfigCenterLoader(addr, prefix)
	if err = loader.Init(); err != nil {
		log.Println(err)
		os.Exit(0)
	}
	log.Println("Success")
	var v *config.Value

	if v, err = loader.Get(config.ROOT); err != nil {
		log.Printf("Get Root error: %v\n", err)
		os.Exit(0)
	}
	if v == nil {
		log.Printf("Get Root not exists\n")
		os.Exit(0)
	}

	log.Println(v)

	var m = &myconfig{}
	if err = v.Populate(&m); err != nil {
		log.Printf("Populate error: %v\n", err)
		os.Exit(0)
	}
	log.Println(m)
	/*
		if v, err = loader.Get("mysql"); err != nil {
			log.Printf("Get \"mysql\" error: %v\n", err)
			os.Exit(0)
		}
		if v == nil {
			log.Println("\"mysql\" not found")
			os.Exit(0)
		}
		var m1 interface{}
		if err = v.Populate(&m1); err != nil {
			log.Println("Populate error %v\n", err)
			os.Exit(0)
		}
		log.Println(m1)

		if v, err = loader.Get("name"); err != nil {
			log.Printf("Get \"mysql.user\" error: %v\n", err)
			os.Exit(0)
		}
		if v == nil {
			log.Println("\"mysql.user\" not found")
			os.Exit(0)
		}

		fmt.Println(v)
		fmt.Printf("mysql.user(string)=%v\n", v.AsString())*/
	/*
		var err error
		var data []byte

		data, err = ioutil.ReadFile("a.yaml")
		if err != nil {
			panic(err)
		}

		provider := config.NewYamlProvider()
		err = provider.Init(data)
		if err != nil {
			panic(err)
		}

		// Test ROOT
		var v *config.Value
		v, err = provider.Get(config.ROOT)
		if err != nil {
			panic(err)
		}
		if v == nil {
			log.Printf("Get ROOT fail")
		} else {

			var m interface{}
			err = v.Populate(&m)
			if err != nil {
				panic(err)
			}
			log.Println(m)
		}

		v, err = provider.Get("server.name")
		if err != nil {
			panic(err)
		}
		if v == nil {
			log.Println("server.name not found")
		} else {
			log.Printf("server.name=%v\n", v.AsString())
		}
	*/
}
