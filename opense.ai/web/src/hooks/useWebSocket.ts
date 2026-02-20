"use client";

import { useEffect, useRef, useCallback } from "react";
import { useStore } from "@/store";
import { getMarketWS, getChatWS, getAlertWS, type WebSocketClient } from "@/lib/ws";
import type { WSMessage, Quote, ChatMessage, Alert } from "@/lib/types";

type WebSocketChannel = "market" | "chat" | "alerts";

interface UseWebSocketOptions {
  channels?: WebSocketChannel[];
  tickers?: string[];
  autoConnect?: boolean;
}

export function useWebSocket(options: UseWebSocketOptions = {}) {
  const {
    channels = ["market"],
    tickers = [],
    autoConnect = true,
  } = options;

  const {
    setQuote,
    watchlist,
    addMessage,
    setStreaming,
    setAlerts,
    alerts,
  } = useStore();

  const wsRefs = useRef<Map<WebSocketChannel, WebSocketClient>>(new Map());
  const cleanupRefs = useRef<Array<() => void>>([]);

  const handleMarketMessage = useCallback(
    (msg: WSMessage) => {
      if (msg.type === "quote" && msg.data) {
        const quote = msg.data as Quote;
        if (quote.ticker) {
          setQuote(quote.ticker, quote);
        }
      }
    },
    [setQuote]
  );

  const handleChatMessage = useCallback(
    (msg: WSMessage) => {
      if (msg.type === "chat_message" && msg.data) {
        const chatMsg = msg.data as ChatMessage;
        addMessage(chatMsg);
      } else if (msg.type === "stream_start") {
        setStreaming(true);
      } else if (msg.type === "stream_end") {
        setStreaming(false);
      }
    },
    [addMessage, setStreaming]
  );

  const handleAlertMessage = useCallback(
    (msg: WSMessage) => {
      if (msg.type === "alert_triggered" && msg.data) {
        const alert = msg.data as Alert;
        setAlerts(
          alerts.map((a) =>
            a.id === alert.id ? { ...a, status: "triggered" as const, triggeredAt: alert.triggeredAt, value: alert.value } : a
          )
        );
        // Browser notification
        if ("Notification" in window && Notification.permission === "granted") {
          new Notification("Alert Triggered", {
            body: `${alert.expression} — Value: ${alert.value ?? "N/A"}`,
            icon: "/favicon.ico",
          });
        }
      }
    },
    [alerts, setAlerts]
  );

  useEffect(() => {
    if (!autoConnect) return;

    const cleanups: Array<() => void> = [];

    if (channels.includes("market")) {
      const ws = getMarketWS();
      wsRefs.current.set("market", ws);
      ws.connect();
      const unsub = ws.on("*", handleMarketMessage);
      cleanups.push(unsub);

      // Subscribe to watchlist tickers
      const allTickers = [...new Set([...watchlist, ...tickers])];
      allTickers.forEach((ticker) => {
        ws.subscribe(`quote:${ticker}`);
      });
    }

    if (channels.includes("chat")) {
      const ws = getChatWS();
      wsRefs.current.set("chat", ws);
      ws.connect();
      const unsub = ws.on("*", handleChatMessage);
      cleanups.push(unsub);
    }

    if (channels.includes("alerts")) {
      const ws = getAlertWS();
      wsRefs.current.set("alerts", ws);
      ws.connect();
      const unsub = ws.on("*", handleAlertMessage);
      cleanups.push(unsub);
    }

    cleanupRefs.current = cleanups;

    return () => {
      cleanups.forEach((fn) => fn());
    };
  }, [
    autoConnect,
    channels.join(","),
    watchlist.join(","),
    tickers.join(","),
    handleMarketMessage,
    handleChatMessage,
    handleAlertMessage,
  ]);

  // Re-subscribe when watchlist changes (market channel)
  useEffect(() => {
    const ws = wsRefs.current.get("market");
    if (!ws) return;
    const allTickers = [...new Set([...watchlist, ...tickers])];
    allTickers.forEach((ticker) => {
      ws.subscribe(`quote:${ticker}`);
    });
  }, [watchlist, tickers]);

  const sendChatMessage = useCallback((content: string) => {
    const ws = wsRefs.current.get("chat");
    if (ws) {
      ws.send({ type: "chat_message", data: { content } });
    }
  }, []);

  const requestNotificationPermission = useCallback(async () => {
    if ("Notification" in window && Notification.permission === "default") {
      await Notification.requestPermission();
    }
  }, []);

  return {
    sendChatMessage,
    requestNotificationPermission,
    isConnected: (channel: WebSocketChannel) => {
      const ws = wsRefs.current.get(channel);
      return ws?.isConnected ?? false;
    },
  };
}
