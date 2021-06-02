package huobi

type MessageBase struct {
	Action string `json:"action"`
}

type MessagePing struct {
	MessageBase
	Data MessagePingBody `json:"data"`
}

type MessagePingBody struct {
	TS uint64 `json:"ts"`
}
