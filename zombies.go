package zombies

import "errors"

var (
	// ErrZombieDoesntExist is an error which says a Zombie does not exist
	ErrZombieDoesntExist = errors.New("Zombie does not exist")
)

// Zombies represent a pool of zombies
type Zombies struct {
	pool map[int64]*Zombie
}

// New creates a new pool of Zombies
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
func (zs *Zombies) Revive(id int64) (*Zombie, error) {
	if !zs.Exists(id) {
		return nil, ErrZombieDoesntExist
	}

	z := zs.pool[id]

	return z, nil
}

// New creates a new zombie and adds it to the pool
func (zs *Zombies) New(id int64, server, nick string) (*Zombie, error) {
	z, err := NewZombie(server, nick)
	if err != nil {
		return nil, err
	}

	zs.pool[id] = z

	return z, nil
}
