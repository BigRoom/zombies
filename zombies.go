package zombies

import (
	"errors"

	"github.com/gorilla/websocket"
)

var (
	ErrZombieDoesntExist = errors.New("Zombie does not exist")
)

// A pool of zombies
type Zombies struct {
	pool map[int64]*Zombie
}

func New() *Zombies {
	return &Zombies{
		pool: make(map[int64]*Zombie),
	}
}

// Exists checks whether a zombie has been created
func (zs *Zombies) Exists(id int64) bool {
	return zs.pool[id] != nil
}

// Revive retrieves a zombie from the pool and associates it with a new WebSocket
func (zs *Zombies) Revive(id int64, c websocket.Conn) (*Zombie, error) {
	if !zs.Exists(id) {
		return nil, ErrZombieDoesntExist
	}

	z := zs.pool[id]

	*z.WSConn = c

	return z, nil
}

// New creates a new zombie and adds it to the pool
func (zs *Zombies) New(id int64, server, nick string, c *websocket.Conn) (*Zombie, error) {
	z, err := NewZombie(server, nick, c)
	if err != nil {
		return nil, err
	}

	zs.pool[id] = z

	return z, nil
}
