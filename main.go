package main

import (
	"log"
	"os"
	"os/signal"
	"poker-engine/engine"
	"poker-engine/server"
	"syscall"
)

func main() {
	tableManager := engine.NewTableManager()
	tcpServer := server.NewTCPServer(":8080", tableManager)

	go func() {
		log.Println("🎰 Poker engine starting on :8080")
		if err := tcpServer.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	tcpServer.Stop()
}
