// Command pool is a kite which represents cluster of zombies
package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/bigroom/zombies"
	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"
	"github.com/koding/kite"
	"github.com/paked/configure"
)

var (
	pool    *zombies.Zombies
	eClient client.Client
	store   client.KeysAPI

	id int64

	conf = configure.New()
	port = conf.Int("port", 5555, "THe port you want to listen on")
)

func main() {
	conf.Use(configure.NewEnvironment())
	conf.Use(configure.NewFlag())

	conf.Parse()

	rand.Seed(time.Now().UnixNano())
	id = rand.Int63()

	var err error

	ip := os.Getenv("STORE_PORT_4001_TCP_ADDR")
	log.Println("Got ip", ip)
	cfg := client.Config{
		Endpoints: []string{"http://" + ip + ":4001"},
		Transport: client.DefaultTransport,
	}

	eClient, err = client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	store = client.NewKeysAPI(eClient)

	pool = zombies.New()

	k := kite.New("pool", "1.0.0")

	k.Config.Port = *port
	k.Config.IP = "0.0.0.0"

	k.HandleFunc("add", addZombie).
		DisableAuthentication()

	k.HandleFunc("send", sendZombie).
		DisableAuthentication()

	k.HandleFunc("join", joinZombie).
		DisableAuthentication()

	go k.Run()

	<-k.ServerReadyNotify()

	fmt.Println("Serving on port", k.Port, "provided", k.Config.Port)

	<-k.ServerCloseNotify()
}

// addZombie adds a new zombie to the runnning pool. It takes a zombies.Add struct and returns the port
func addZombie(r *kite.Request) (interface{}, error) {
	// Write to etcd
	add := zombies.Add{}
	r.Args.One().MustUnmarshal(&add)

	z, err := pool.New(add.ID, add.Server, add.Nick)
	if err != nil {
		return z, err
	}

	// tell etcd that the zombie is in this pool
	resp, err := store.Set(context.Background(), fmt.Sprintf("/zombies/%v", add.ID), fmt.Sprintf("%v", *port), nil)
	if err != nil {
		return z, err
	}

	log.Printf("Setting is done. Here is the metadata %v", resp)

	return 3001, nil
}

func joinZombie(r *kite.Request) (interface{}, error) {
	join := zombies.Join{}
	r.Args.One().MustUnmarshal(&join)

	z, err := pool.Revive(join.ID)
	if err != nil {
		return z, err
	}

	z.Join(join.Channel)

	return 3001, nil
}

func sendZombie(r *kite.Request) (interface{}, error) {
	send := zombies.Send{}
	r.Args.One().MustUnmarshal(&send)

	z, err := pool.Revive(send.ID)
	if err != nil {
		return z, err
	}

	log.Println(z)

	z.Messages <- send

	return nil, nil
}
