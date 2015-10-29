package zombies

type Add struct {
	ID     int64  `json:"id"`
	Nick   string `json:"nick"`
	Server string `json:"server"`
}

type Join struct {
	ID      int64  `json:"id"`
	Channel string `json:"channel"`
}

type Send struct {
	ID      int64  `json:"id"`
	Channel string `json:"channel"`
	Message string `json:"message"`
}

type Channels struct {
	Channels []string `json:"channels"`
}
