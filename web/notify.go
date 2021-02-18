package web

import "github.com/gorilla/websocket"

// NextActionNotify : f
func NextActionNotify(tp, value string, c *websocket.Conn) {
	c.WriteJSON(FlowData{
		Tp:    "Notify",
		Id:    tp,
		Value: value,
	})
}
