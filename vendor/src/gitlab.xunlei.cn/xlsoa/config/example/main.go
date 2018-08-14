package main

import (
	"flag"
	"fmt"
	xlsoa_config "gitlab.xunlei.cn/xlsoa/config"
	"log"
	"os"
)

type config struct {
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

func (c *config) String() string {
	ss := fmt.Sprintf("Config{ ")
	ss += fmt.Sprintf(" Name: %v, ", c.Name)
	ss += fmt.Sprintf(" Server{ addr: %v, port: %v }", c.Server.Addr, c.Server.Port)
	ss += fmt.Sprintf(", Mysql{ host: %v, port: %v, user: %v, password: %v}, ",
		c.Mysql.Host, c.Mysql.Port, c.Mysql.User, c.Mysql.Password)
	ss += fmt.Sprintf(" Token{ ttl: %v } ", c.Token.Ttl)
	ss += " }"
	return ss
}

func main() {
	var err error

	var (
		help     = false
		addr     = ""
		dc       = ""
		node     = ""
		instance = ""
		name     = ""
	)

	flag.BoolVar(&help, "h", false, "Help message.")
	flag.StringVar(&addr, "addr", "", "Config server addr")
	flag.StringVar(&dc, "dc", "", "Datacenter name")
	flag.StringVar(&node, "node", "", "Node name")
	flag.StringVar(&instance, "instance", "", "Instance name")
	flag.StringVar(&name, "name", "", "Service name")
	flag.Parse()

	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if name == "" {
		log.Println("Name is required!\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var c *xlsoa_config.Config
	var loader xlsoa_config.Loader
	var v *xlsoa_config.Value

	// Init options
	opts := []xlsoa_config.OptionFunc{}
	if addr != "" {
		opts = append(opts, xlsoa_config.WithAddr(addr))
	}
	if dc != "" {
		opts = append(opts, xlsoa_config.WithDcName(dc))
	}
	if node != "" {
		opts = append(opts, xlsoa_config.WithNodeName(node))
	}
	if instance != "" {
		opts = append(opts, xlsoa_config.WithInstanceName(instance))
	}

	// New & Load
	c = xlsoa_config.New(name, opts...)
	if loader, err = c.Load(); err != nil {
		log.Fatal(err)
	}

	// Watch
	// We MUST get ch before Get, avoid missing update events between Get and Watch.
	var ch chan bool
	ch, err = loader.Watch(xlsoa_config.ROOT)
	if err != nil {
		log.Fatal(err)
	}

	// Get
	if v, err = loader.Get(xlsoa_config.ROOT); err != nil {
		log.Fatal(err)
	}
	if v == nil {
		log.Fatal("Not exists")
	}

	var config = &config{}
	if err = v.Populate(config); err != nil {
		log.Fatalf("Populate error: %v\n", err)
	}
	log.Println("Config loaded:")
	log.Println(config)

	// Deal update events
	for {
		select {
		case <-ch:
			v, err = loader.Get(xlsoa_config.ROOT)
			if err != nil || v == nil {
				continue
			}

			err = v.Populate(config)
			if err != nil {
				log.Fatalf("Updated config Populate error: %v\n", err)
				break
			}
			log.Println("Config updated:")
			log.Println(config)

		}
	}

}
