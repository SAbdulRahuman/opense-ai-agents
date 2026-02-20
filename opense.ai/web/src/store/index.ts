// ============================================================================
// OpeNSE.ai — Global Zustand Store (combined slices)
// ============================================================================

import { create } from "zustand";
import { createChatSlice, type ChatSlice } from "./chatSlice";
import { createMarketSlice, type MarketSlice } from "./marketSlice";
import { createQuerySlice, type QuerySlice } from "./querySlice";

export type AppStore = ChatSlice & MarketSlice & QuerySlice;

export const useStore = create<AppStore>()((...a) => ({
  ...createChatSlice(...a),
  ...createMarketSlice(...a),
  ...createQuerySlice(...a),
}));
