package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"poker-engine/engine"
	"poker-engine/models"
	"sync"
)

type TCPServer struct {
	address      string
	listener     net.Listener
	handler      *CommandHandler
	tableManager *engine.TableManager
	conn         net.Conn
	mu           sync.Mutex
	stopChan     chan struct{}
}

func NewTCPServer(address string, tableManager *engine.TableManager) *TCPServer {
	return &TCPServer{
		address:      address,
		handler:      NewCommandHandler(tableManager),
		tableManager: tableManager,
		stopChan:     make(chan struct{}),
	}
}

func (s *TCPServer) Start() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}

	s.listener = listener
	log.Printf("TCP server listening on %s", s.address)

	go s.eventBroadcaster()

	for {
		select {
		case <-s.stopChan:
			return nil
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				log.Printf("Error accepting connection: %v", err)
				continue
			}

			log.Printf("Client connected from %s", conn.RemoteAddr().String())
			s.mu.Lock()
			s.conn = conn
			s.mu.Unlock()

			go s.handleConnection(conn)
		}
	}
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.mu.Lock()
		s.conn = nil
		s.mu.Unlock()
		log.Printf("Client disconnected")
	}()

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		
		var cmd models.Command
		if err := json.Unmarshal(line, &cmd); err != nil {
			response := models.Response{
				Success: false,
				Error:   fmt.Sprintf("invalid JSON: %v", err),
			}
			s.sendResponse(conn, response)
			continue
		}

		response := s.handler.Handle(cmd)
		s.sendResponse(conn, response)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Scanner error: %v", err)
	}
}

func (s *TCPServer) sendResponse(conn net.Conn, response models.Response) {
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return
	}

	data = append(data, '\n')
	
	_, err = conn.Write(data)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func (s *TCPServer) sendEvent(event models.Event) {
	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()

	if conn == nil {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	data = append(data, '\n')
	
	_, err = conn.Write(data)
	if err != nil {
		log.Printf("Error writing event: %v", err)
	}
}

func (s *TCPServer) eventBroadcaster() {
	eventChan := s.tableManager.GetEventChannel()
	for {
		select {
		case <-s.stopChan:
			return
		case event := <-eventChan:
			s.sendEvent(event)
		}
	}
}

func (s *TCPServer) Stop() {
	close(s.stopChan)
	if s.listener != nil {
		s.listener.Close()
	}
	s.mu.Lock()
	if s.conn != nil {
		s.conn.Close()
	}
	s.mu.Unlock()
}
