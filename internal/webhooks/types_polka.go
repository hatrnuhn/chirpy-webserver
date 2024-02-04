package webhooks

type PolkaReq struct {
	Event string
	Data  map[string]int
}
