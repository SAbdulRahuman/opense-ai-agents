// ============================================================================
// OpeNSE.ai — WebSocket Client (reconnecting WS for streaming data)
// ============================================================================

import type { WSMessage } from "./types";

type WSCallback = (msg: WSMessage) => void;

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectDelay = 1000;
  private heartbeatInterval: ReturnType<typeof setInterval> | null = null;
  private listeners: Map<string, Set<WSCallback>> = new Map();
  private subscriptions: Set<string> = new Set();
  private isClosing = false;

  constructor(path: string = "") {
    const base = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080/api/v1/ws";
    this.url = `${base}${path}`;
  }

  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) return;
    this.isClosing = false;

    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        this.reconnectAttempts = 0;
        this.startHeartbeat();
        // Re-subscribe to previously subscribed topics
        this.subscriptions.forEach((topic) => {
          this.send({ type: "subscribe", data: { topic } });
        });
        this.emit("connected", { type: "connected", data: null });
      };

      this.ws.onmessage = (event) => {
        try {
          const msg: WSMessage = JSON.parse(event.data);
          this.emit(msg.type, msg);
          this.emit("*", msg); // wildcard listeners
        } catch {
          console.error("Failed to parse WS message:", event.data);
        }
      };

      this.ws.onclose = () => {
        this.stopHeartbeat();
        this.emit("disconnected", { type: "disconnected", data: null });
        if (!this.isClosing) {
          this.attemptReconnect();
        }
      };

      this.ws.onerror = (error) => {
        console.error("WebSocket error:", error);
        this.emit("error", { type: "error", data: error });
      };
    } catch (error) {
      console.error("Failed to create WebSocket:", error);
      this.attemptReconnect();
    }
  }

  disconnect(): void {
    this.isClosing = true;
    this.stopHeartbeat();
    this.ws?.close();
    this.ws = null;
  }

  send(msg: WSMessage): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg));
    }
  }

  subscribe(topic: string): void {
    this.subscriptions.add(topic);
    this.send({ type: "subscribe", data: { topic } });
  }

  unsubscribe(topic: string): void {
    this.subscriptions.delete(topic);
    this.send({ type: "unsubscribe", data: { topic } });
  }

  on(type: string, callback: WSCallback): () => void {
    if (!this.listeners.has(type)) {
      this.listeners.set(type, new Set());
    }
    this.listeners.get(type)!.add(callback);

    // Return unsubscribe function
    return () => {
      this.listeners.get(type)?.delete(callback);
    };
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  private emit(type: string, msg: WSMessage): void {
    this.listeners.get(type)?.forEach((cb) => cb(msg));
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error("Max reconnection attempts reached");
      return;
    }

    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts);
    this.reconnectAttempts++;

    setTimeout(() => {
      this.connect();
    }, Math.min(delay, 30000)); // Cap at 30 seconds
  }

  private startHeartbeat(): void {
    this.stopHeartbeat();
    this.heartbeatInterval = setInterval(() => {
      this.send({ type: "ping", data: null });
    }, 30000);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }
}

// Singleton instances for different WS channels
let marketWS: WebSocketClient | null = null;
let chatWS: WebSocketClient | null = null;
let alertWS: WebSocketClient | null = null;

export function getMarketWS(): WebSocketClient {
  if (!marketWS) {
    marketWS = new WebSocketClient("/market");
  }
  return marketWS;
}

export function getChatWS(): WebSocketClient {
  if (!chatWS) {
    chatWS = new WebSocketClient("/chat");
  }
  return chatWS;
}

export function getAlertWS(): WebSocketClient {
  if (!alertWS) {
    alertWS = new WebSocketClient("/alerts");
  }
  return alertWS;
}
