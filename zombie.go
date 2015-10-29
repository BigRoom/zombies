package zombies

import (
	"fmt"
	"log"

	"github.com/nickvanw/ircx"
	"github.com/sorcix/irc"
)

// Zombie is a struct which represents a Big Room user inside of an IRC server.
type Zombie struct {
	Messages chan Send
	irc      *ircx.Bot
	nick     string
	server   string
	Channels []string
}

// NewZombie either creates or retrieves a zombie. Zombies will be created if
// they do not already exist for that user on a server. Zombies will be
// retrieved if a zombie for the user on a server already exists
func NewZombie(server, nick string) (*Zombie, error) {
	zombie := ircx.Classic(server, nick)

	if err := zombie.Connect(); err != nil {
		return nil, err
	}

	user := &Zombie{
		Messages: make(chan Send),
		irc:      zombie,
		server:   server,
		nick:     nick,
	}

	zombie.HandleFunc(irc.PING, user.pingHandler)
	zombie.HandleFunc(irc.RPL_WELCOME, user.registerHandler)
	zombie.HandleFunc(irc.JOIN, user.messageHandler)

	go zombie.HandleLoop()

	return user, nil
}

func (z *Zombie) SetNick(name string) {
	z.nick = name

	z.irc.Sender.Send(&irc.Message{
		Command: irc.NICK,
		Params:  []string{z.nick},
	})
}

// Join makes the zombie join a bunch of channels. Channels are parsed in the form: x.x.x.x:6667/#channel
func (z *Zombie) Join(channels ...string) {
	// uggerrsss
L:
	for i := range channels {
		channel, err := ParseChannelKey(channels[i])
		if err != nil {
			log.Println("Failed getting key:", err)
		}

		channels[i] = channel

		for _, c := range z.Channels {
			if c == channels[i] {
				break L
			}
		}
	}

	fmt.Println("Translated channels:", channels)

	z.irc.Sender.Send(&irc.Message{
		Command: irc.JOIN,
		Params:  channels,
	})

	z.Channels = append(z.Channels, channels...)
}

func (z *Zombie) messageHandler(s ircx.Sender, m *irc.Message) {
	go func() {
		for {
			log.Println("Waiting for message")
			msg := <-z.Messages

			log.Printf("Got message '%v'. Sending to IRC on channel '%v'...", msg.Message, msg.Channel)

			channel, err := ParseChannelKey(msg.Channel)
			if err != nil {
				log.Println("Couldnt get channel:", err)
				return
			}

			log.Printf("Got channel (%v)", channel)

			err = s.Send(&irc.Message{
				Command:  irc.PRIVMSG,
				Params:   []string{channel},
				Trailing: msg.Message,
			})

			if err != nil {
				log.Println("Couldn't send message:", err)
			}

			log.Println("Message sent")
		}
	}()
}

func (z *Zombie) registerHandler(s ircx.Sender, m *irc.Message) {
	log.Println("Registering")
}

func (z *Zombie) pingHandler(s ircx.Sender, m *irc.Message) {
	s.Send(&irc.Message{
		Command:  irc.PONG,
		Params:   m.Params,
		Trailing: m.Trailing,
	})
}

func ParseChannelKey(s string) (string, error) {
	var channel string

	var (
		h1 int
		h2 int
		h3 int
		h4 int
	)

	_, err := fmt.Sscanf(s, "%d.%d.%d.%d:6667/%v", &h1, &h2, &h3, &h4, &channel)
	if err != nil {
		log.Println("failing!")
		return channel, err
	}

	return channel, nil
}
