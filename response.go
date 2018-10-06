package main

import (
	"log"

	"golang.org/x/net/websocket"
)

type analyzeResponseStatus int

const (
	statusSuccess analyzeResponseStatus = iota
	statusFailure
	statusComplete
)

type analyzeResponse struct {
	Result string
	Status analyzeResponseStatus
}

// ResponseSuccess returns success response to client.
func ResponseSuccess(ws *websocket.Conn, message string) {
	writeResponse(ws, message, statusSuccess)
}

// ResponseFailure returns failure response to client.
func ResponseFailure(ws *websocket.Conn, message string) {
	writeResponse(ws, message, statusFailure)
}

// ResponseComplete returns complete response to client.
func ResponseComplete(ws *websocket.Conn, message string) {
	writeResponse(ws, message, statusComplete)
}

func writeResponse(ws *websocket.Conn, message string, status analyzeResponseStatus) {
	if err := websocket.JSON.Send(ws, analyzeResponse{Result: message, Status: status}); err != nil {
		log.Printf("couldn't send websocket response %v", err)
	}
}
