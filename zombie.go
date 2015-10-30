package zombies

import (
	"fmt"
	log "github.com/sirupsen/logrus"

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
	n := []string{}

	for i := range channels {
		channel, err := ParseChannelKey(channels[i])
		if err != nil {
			log.WithFields(log.Fields{
				"channel": channel,
				"error":   err,
			}).Error("Failed getting key")
		}

		add := true
		for _, c := range z.Channels {
			if c == channel {
				add = false
				break
			}
		}

		if add {
			n = append(n, channel)
		}
	}

	log.WithFields(log.Fields{
		"channels": n,
	}).Info("Added channels")

	z.irc.Sender.Send(&irc.Message{
		Command: irc.JOIN,
		Params:  channels,
	})

	z.Channels = append(z.Channels, n...)
}

func (z *Zombie) messageHandler(s ircx.Sender, m *irc.Message) {
	go func() {
		for {
			log.Info("Waiting for message")
			msg := <-z.Messages

			log.WithFields(log.Fields{
				"message":     msg.Message,
				"channel_key": msg.Channel,
			}).Info("Got message")

			channel, err := ParseChannelKey(msg.Channel)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Couldnt get channel")
				return
			}

			log.WithFields(log.Fields{
				"channel": channel,
			}).Info("Got channel")

			err = s.Send(&irc.Message{
				Command:  irc.PRIVMSG,
				Params:   []string{channel},
				Trailing: msg.Message,
			})

			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Warn("Couldn't send message")
			}

			log.Info("Message sent")
		}
	}()
}

func (z *Zombie) registerHandler(s ircx.Sender, m *irc.Message) {
	log.Info("Registered")
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
		return channel, err
	}

	return channel, nil
}
