// ============================================================================
// OpeNSE.ai — Trading Store Slice (Zustand)
// Manages orders, positions, funds, order window state, and watchlist tabs.
// ============================================================================

import type { StateCreator } from "zustand";
import type {
  Order,
  OrderRequest,
  OrderResponse,
  Position,
  Margins,
  OrderSide,
} from "@/lib/types";

export interface TradingSlice {
  // State
  orders: Order[];
  positions: Position[];
  margins: Margins | null;
  isOrderWindowOpen: boolean;
  orderWindowTicker: string;
  orderWindowSide: OrderSide;
  isPlacingOrder: boolean;
  activeWatchlistTab: number;
  ordersLoading: boolean;
  positionsLoading: boolean;
  fundsLoading: boolean;

  // Actions
  fetchOrders: () => Promise<void>;
  fetchPositions: () => Promise<void>;
  fetchFunds: () => Promise<void>;
  submitOrder: (req: OrderRequest) => Promise<OrderResponse | null>;
  cancelOrderById: (id: string) => Promise<boolean>;
  openOrderWindow: (ticker: string, side: OrderSide) => void;
  closeOrderWindow: () => void;
  setActiveWatchlistTab: (tab: number) => void;
}

export const createTradingSlice: StateCreator<TradingSlice> = (set, get) => ({
  orders: [],
  positions: [],
  margins: null,
  isOrderWindowOpen: false,
  orderWindowTicker: "",
  orderWindowSide: "BUY",
  isPlacingOrder: false,
  activeWatchlistTab: 0,
  ordersLoading: false,
  positionsLoading: false,
  fundsLoading: false,

  fetchOrders: async () => {
    set({ ordersLoading: true });
    try {
      const { getOrders } = await import("@/lib/api");
      const orders = await getOrders();
      set({ orders });
    } catch {
      // ignore
    } finally {
      set({ ordersLoading: false });
    }
  },

  fetchPositions: async () => {
    set({ positionsLoading: true });
    try {
      const { getPositions } = await import("@/lib/api");
      const positions = await getPositions();
      set({ positions });
    } catch {
      // ignore
    } finally {
      set({ positionsLoading: false });
    }
  },

  fetchFunds: async () => {
    set({ fundsLoading: true });
    try {
      const { getFunds } = await import("@/lib/api");
      const margins = await getFunds();
      set({ margins });
    } catch {
      // ignore
    } finally {
      set({ fundsLoading: false });
    }
  },

  submitOrder: async (req: OrderRequest) => {
    set({ isPlacingOrder: true });
    try {
      const { placeOrder } = await import("@/lib/api");
      const resp = await placeOrder(req);
      // Refresh orders after placing
      get().fetchOrders();
      return resp;
    } catch {
      return null;
    } finally {
      set({ isPlacingOrder: false });
    }
  },

  cancelOrderById: async (id: string) => {
    try {
      const { cancelOrder } = await import("@/lib/api");
      await cancelOrder(id);
      get().fetchOrders();
      return true;
    } catch {
      return false;
    }
  },

  openOrderWindow: (ticker: string, side: OrderSide) =>
    set({ isOrderWindowOpen: true, orderWindowTicker: ticker, orderWindowSide: side }),

  closeOrderWindow: () =>
    set({ isOrderWindowOpen: false, orderWindowTicker: "", orderWindowSide: "BUY" }),

  setActiveWatchlistTab: (tab: number) => set({ activeWatchlistTab: tab }),
});
