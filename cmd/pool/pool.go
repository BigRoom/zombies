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
	"github.com/getsentry/raven-go"
	"github.com/koding/kite"
	"github.com/paked/configure"
)

var (
	pool    *zombies.Zombies
	eClient client.Client
	store   client.KeysAPI
	sentry  *raven.Client

	uid int64

	conf      = configure.New()
	port      = conf.Int("port", 3001, "THe port you want to listen on")
	sentryDSN = conf.String("sentry-dsn", "", "The sentry DSN you want to lose")
)

func main() {
	var err error

	conf.Use(configure.NewEnvironment())
	conf.Use(configure.NewFlag())

	conf.Parse()

	sentry, err = raven.NewClient(*sentryDSN, nil)
	if err != nil {
		log.Println("Could not connect to sentry:", err)
	}

	rand.Seed(time.Now().UnixNano())
	uid = rand.Int63()

	setupETCD()

	pool = zombies.New()

	k := kite.New("pool", "1.0.0")

	k.Config.Port = *port
	k.Config.IP = "0.0.0.0"

	bindHandlers(k)

	go k.Run()

	<-k.ServerReadyNotify()

	fmt.Println("Serving on port", k.Port, "provided", k.Config.Port)

	<-k.ServerCloseNotify()
}

func setupETCD() {
	var err error

	ip := os.Getenv("STORE_PORT_4001_TCP_ADDR")
	log.Println("Got ip", ip)

	cfg := client.Config{
		Endpoints: []string{"http://" + ip + ":4001"},
		Transport: client.DefaultTransport,
	}

	eClient, err = client.New(cfg)
	if err != nil {
		log.Println(err)
		sentry.CaptureErrorAndWait(err, nil)
		return
	}

	store = client.NewKeysAPI(eClient)

	// Yes. This is real.
	store.Set(context.Background(), "/zombies", "", &client.SetOptions{
		Dir: true,
	})
}

func bindHandlers(k *kite.Kite) {
	k.HandleFunc("add", addZombie).
		DisableAuthentication()

	k.HandleFunc("send", sendZombie).
		DisableAuthentication()

	k.HandleFunc("join", joinZombie).
		DisableAuthentication()

	k.HandleFunc("exists", existsZombie).
		DisableAuthentication()

	k.HandleFunc("channels", channelsZombie).
		DisableAuthentication()
}

// addZombie adds a new zombie to the runnning pool. It takes a zombies.Add struct and returns the port
func addZombie(r *kite.Request) (interface{}, error) {
	defer sentry.ClearContext()
	sentry.SetTagsContext(map[string]string{"type": "add"})

	// Write to etcd
	add := zombies.Add{}
	r.Args.One().MustUnmarshal(&add)

	z, err := pool.New(add.ID, add.Server, add.Nick)
	if err != nil {
		sentry.CaptureErrorAndWait(err, nil)
		return z, err
	}

	// tell etcd that the zombie is in this pool
	resp, err := store.Set(context.Background(), fmt.Sprintf("/zombies/%v", add.ID), makeKey(), nil)
	if err != nil {
		sentry.CaptureErrorAndWait(err, nil)
		return z, err
	}

	log.Printf("Setting is done. Here is the metadata %v", resp)

	return *port, nil
}

// existsZombie consults etcd to check if a zombie exists. If a zombie existed in a previous version of the pool
// it will overwritten
func existsZombie(r *kite.Request) (interface{}, error) {
	defer sentry.ClearContext()
	sentry.SetTagsContext(map[string]string{"type": "exists"})

	id := int64(r.Args.One().MustFloat64())

	key := fmt.Sprintf("/zombies/%v", id)
	resp, err := store.Get(context.Background(), key, nil)
	if err != nil {
		log.Println("error: ", err)
		sentry.CaptureErrorAndWait(err, nil)

		return false, nil //zombies.ErrZombieDoesntExist
	}

	if makeKey() == resp.Node.Value {
		log.Println("Zombie does exist!")
		return true, nil
	} else if p, u := translateKey(resp.Node.Value); p == *port && u != uid {
		_, err := store.Delete(context.Background(), key, nil)
		if err != nil {
			sentry.CaptureErrorAndWait(err, nil)
			return false, err
		}

		fmt.Println("Deleted old key")
	}

	log.Println("Zombie does not exist...")

	return false, nil
}

// joinZombie will join a new irc channel
func joinZombie(r *kite.Request) (interface{}, error) {
	defer sentry.ClearContext()
	sentry.SetTagsContext(map[string]string{"type": "join"})

	join := zombies.Join{}
	r.Args.One().MustUnmarshal(&join)

	z, err := pool.Revive(join.ID)
	if err != nil {
		sentry.CaptureErrorAndWait(err, nil)
		return z, err
	}

	z.Join(join.Channel)

	return 3001, nil
}

// sendZombie adds a message to a queue of messages
func sendZombie(r *kite.Request) (interface{}, error) {
	defer sentry.ClearContext()
	sentry.SetTagsContext(map[string]string{"type": "send"})

	send := zombies.Send{}
	r.Args.One().MustUnmarshal(&send)

	z, err := pool.Revive(send.ID)
	if err != nil {
		sentry.CaptureErrorAndWait(err, nil)
		return z, err
	}

	z.Messages <- send

	return nil, nil
}

func channelsZombie(r *kite.Request) (interface{}, error) {
	defer sentry.ClearContext()
	sentry.SetTagsContext(map[string]string{"type": "channels"})

	id := int64(r.Args.One().MustFloat64())

	z, err := pool.Revive(id)
	if err != nil {
		sentry.CaptureErrorAndWait(err, nil)
		return z, err
	}

	return zombies.Channels{
		Channels: z.Channels,
	}, nil
}

// makeKey assembles a key used to save a zombie in etcd
func makeKey() string {
	return fmt.Sprintf("%v:%v", *port, uid)
}

// translateKey takes a key and converts it back into a port an uid
func translateKey(s string) (int, int64) {
	var p int
	var u int64

	fmt.Sscanf(s, "%d:%d", &p, &u)

	return p, u
}
