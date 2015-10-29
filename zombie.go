package zombies

import (
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

func (z *Zombie) Join(channel ...string) {
	z.irc.Sender.Send(&irc.Message{
		Command: irc.JOIN,
		Params:  channel,
	})
}

func (z *Zombie) messageHandler(s ircx.Sender, m *irc.Message) {
	go func() {
		for {
			log.Println("Waiting for message")
			msg := <-z.Messages

			log.Printf("Got message '%v'. Sending to IRC on channel '%v'...", msg.Message, msg.Channel)
			err := s.Send(&irc.Message{
				Command:  irc.PRIVMSG,
				Params:   []string{"#roomtest"},
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
	z.Join("#roomtest")
}

func (z *Zombie) pingHandler(s ircx.Sender, m *irc.Message) {
	s.Send(&irc.Message{
		Command:  irc.PONG,
		Params:   m.Params,
		Trailing: m.Trailing,
	})
}
