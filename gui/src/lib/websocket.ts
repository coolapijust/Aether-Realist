// WebSocket event stream client
import type { AnyCoreEvent } from '@/types/core';

const WS_URL = 'ws://localhost:9880/api/v1/events';

export class EventStream {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectTimer: number | null = null;
  private onEvent: (event: AnyCoreEvent) => void;
  private onConnect: () => void;
  private onDisconnect: () => void;

  constructor(
    onEvent: (event: AnyCoreEvent) => void,
    onConnect: () => void = () => {},
    onDisconnect: () => void = () => {},
    url: string = WS_URL
  ) {
    this.url = url;
    this.onEvent = onEvent;
    this.onConnect = onConnect;
    this.onDisconnect = onDisconnect;
  }

  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) return;

    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      console.log('[EventStream] Connected');
      this.onConnect();
      
      // Send ping every 30s to keep alive
      setInterval(() => {
        this.send({ action: 'ping' });
      }, 30000);
    };

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as AnyCoreEvent;
        this.onEvent(data);
      } catch (err) {
        console.error('[EventStream] Failed to parse event:', err);
      }
    };

    this.ws.onclose = () => {
      console.log('[EventStream] Disconnected');
      this.onDisconnect();
      this.scheduleReconnect();
    };

    this.ws.onerror = (err) => {
      console.error('[EventStream] Error:', err);
      this.ws?.close();
    };
  }

  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.ws?.close();
    this.ws = null;
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) return;
    
    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectTimer = null;
      console.log('[EventStream] Reconnecting...');
      this.connect();
    }, 5000);
  }

  send(data: unknown): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }
}
