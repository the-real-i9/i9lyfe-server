package appTypes

type ClientUser struct {
	Username string
}

type ServerWSMsg struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}
