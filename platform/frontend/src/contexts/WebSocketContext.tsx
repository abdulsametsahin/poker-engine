import React, { createContext, useContext, useState, useEffect, useRef, ReactNode, useCallback } from 'react';
import { WSMessage } from '../types';
import { WEBSOCKET, API } from '../constants';
import { useAuth } from './AuthContext';

type MessageHandler = (message: WSMessage) => void;

interface WebSocketContextType {
  isConnected: boolean;
  lastMessage: WSMessage | null;
  sendMessage: (message: WSMessage) => void;
  addMessageHandler: (type: string, handler: MessageHandler) => () => void;
  removeMessageHandler: (type: string) => void;
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(undefined);

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }
  return context;
};

interface WebSocketProviderProps {
  children: ReactNode;
}

export const WebSocketProvider: React.FC<WebSocketProviderProps> = ({ children }) => {
  const { token, isAuthenticated } = useAuth();
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<WSMessage | null>(null);

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttemptRef = useRef(0);
  const messageHandlersRef = useRef<Map<string, MessageHandler[]>>(new Map());
  const heartbeatIntervalRef = useRef<NodeJS.Timeout | null>(null);

  const getReconnectDelay = useCallback(() => {
    const delay = Math.min(
      WEBSOCKET.RECONNECT_BACKOFF_MULTIPLIER ** reconnectAttemptRef.current * 1000,
      30000
    );
    return delay;
  }, []);

  const clearHeartbeat = useCallback(() => {
    if (heartbeatIntervalRef.current) {
      clearInterval(heartbeatIntervalRef.current);
      heartbeatIntervalRef.current = null;
    }
  }, []);

  const startHeartbeat = useCallback(() => {
    clearHeartbeat();
    heartbeatIntervalRef.current = setInterval(() => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({ type: 'ping' }));
      }
    }, WEBSOCKET.HEARTBEAT_INTERVAL);
  }, [clearHeartbeat]);

  const connect = useCallback(() => {
    if (!isAuthenticated || !token) {
      return;
    }

    // Don't reconnect if already connected or connecting
    if (wsRef.current?.readyState === WebSocket.OPEN || wsRef.current?.readyState === WebSocket.CONNECTING) {
      return;
    }

    try {
      const wsUrl = API.BASE_URL.replace('http://', 'ws://').replace('https://', 'wss://').replace('/api', '');
      const ws = new WebSocket(`${wsUrl}/ws?token=${token}`);

      ws.onopen = () => {
        console.log('WebSocket connected');
        setIsConnected(true);
        reconnectAttemptRef.current = 0;
        startHeartbeat();
      };

      ws.onclose = () => {
        console.log('WebSocket disconnected');
        setIsConnected(false);
        clearHeartbeat();

        // Attempt to reconnect if authenticated
        if (isAuthenticated && reconnectAttemptRef.current < WEBSOCKET.RECONNECT_ATTEMPTS) {
          const delay = getReconnectDelay();
          console.log(`Reconnecting in ${delay}ms (attempt ${reconnectAttemptRef.current + 1}/${WEBSOCKET.RECONNECT_ATTEMPTS})`);

          reconnectTimeoutRef.current = setTimeout(() => {
            reconnectAttemptRef.current++;
            connect();
          }, delay);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };

      ws.onmessage = (event) => {
        try {
          const message: WSMessage = JSON.parse(event.data);
          setLastMessage(message);

          // Call all registered handlers for this message type
          const handlers = messageHandlersRef.current.get(message.type);
          if (handlers && handlers.length > 0) {
            handlers.forEach(handler => {
              try {
                handler(message);
              } catch (error) {
                console.error(`Error in handler for message type "${message.type}":`, error);
              }
            });
          }
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      wsRef.current = ws;
    } catch (error) {
      console.error('Failed to create WebSocket:', error);
    }
  }, [isAuthenticated, token, getReconnectDelay, startHeartbeat, clearHeartbeat]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    clearHeartbeat();

    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    setIsConnected(false);
  }, [clearHeartbeat]);

  const sendMessage = useCallback((message: WSMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected. Message not sent:', message);
    }
  }, []);

  const addMessageHandler = useCallback((type: string, handler: MessageHandler) => {
    const handlers = messageHandlersRef.current.get(type) || [];
    handlers.push(handler);
    messageHandlersRef.current.set(type, handlers);

    // Return cleanup function to remove this specific handler
    return () => {
      const currentHandlers = messageHandlersRef.current.get(type) || [];
      const filtered = currentHandlers.filter(h => h !== handler);
      if (filtered.length === 0) {
        messageHandlersRef.current.delete(type);
      } else {
        messageHandlersRef.current.set(type, filtered);
      }
    };
  }, []);

  const removeMessageHandler = useCallback((type: string) => {
    messageHandlersRef.current.delete(type);
  }, []);

  // Connect when authenticated
  useEffect(() => {
    if (isAuthenticated) {
      connect();
    } else {
      disconnect();
    }

    return () => {
      disconnect();
    };
  }, [isAuthenticated, connect, disconnect]);

  return (
    <WebSocketContext.Provider
      value={{
        isConnected,
        lastMessage,
        sendMessage,
        addMessageHandler,
        removeMessageHandler,
      }}
    >
      {children}
    </WebSocketContext.Provider>
  );
};
