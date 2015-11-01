package zombies

// Add is the protocol definition used to add a new IRC zombie
type Add struct {
	ID     int64  `json:"id"`
	Nick   string `json:"nick"`
	Server string `json:"server"`
}

// Join is the protocol definition used to make an IRC zombie join a channel
type Join struct {
	ID      int64  `json:"id"`
	Channel string `json:"channel"`
}

// Send is the protocol definition used to make an IRC zombie send a message to the IRC server
type Send struct {
	ID      int64  `json:"id"`
	Channel string `json:"channel"`
	Message string `json:"message"`
}

// Channels is the protocol definition of a response containing a list of channels
type Channels struct {
	Channels []string `json:"channels"`
}
