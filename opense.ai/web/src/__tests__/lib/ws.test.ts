import { describe, it, expect, vi, beforeEach } from "vitest";
import { WebSocketClient } from "@/lib/ws";

// Mock WebSocket
class MockWebSocket {
  static CONNECTING = 0;
  static OPEN = 1;
  static CLOSING = 2;
  static CLOSED = 3;

  readyState = MockWebSocket.OPEN;
  onopen: ((ev: Event) => void) | null = null;
  onclose: ((ev: CloseEvent) => void) | null = null;
  onmessage: ((ev: MessageEvent) => void) | null = null;
  onerror: ((ev: Event) => void) | null = null;
  send = vi.fn();
  close = vi.fn();

  constructor(public url: string) {
    setTimeout(() => this.onopen?.(new Event("open")), 0);
  }
}

Object.defineProperty(global, "WebSocket", { value: MockWebSocket });

describe("WebSocketClient", () => {
  let client: WebSocketClient;

  beforeEach(() => {
    client = new WebSocketClient("ws://localhost:8080/ws");
  });

  it("creates instance", () => {
    expect(client).toBeDefined();
  });

  it("tracks subscriptions", () => {
    client.connect();
    client.subscribe("quote:RELIANCE");
    client.subscribe("quote:TCS");
    // No public API to check subscriptions, but no error
  });

  it("registers message listeners", () => {
    const handler = vi.fn();
    const unsub = client.on("message", handler);
    expect(typeof unsub).toBe("function");
  });

  it("unsubscribes message listeners", () => {
    const handler = vi.fn();
    const unsub = client.on("message", handler);
    unsub();
    // No error after unsubscribe
  });

  it("sends data", () => {
    client.connect();
    // Wait for connection
    return new Promise<void>((resolve) => {
      setTimeout(() => {
        client.send({ type: "test", data: {} });
        resolve();
      }, 10);
    });
  });
});
