// ============================================================================
// OpeNSE.ai — Global Zustand Store (combined slices)
// ============================================================================

import { create } from "zustand";
import { createChatSlice, type ChatSlice } from "./chatSlice";
import { createChartDrawingsSlice, type ChartDrawingsSlice } from "./chartDrawingsSlice";
import { createConfigSlice, type ConfigSlice } from "./configSlice";
import { createMarketSlice, type MarketSlice } from "./marketSlice";
import { createQuerySlice, type QuerySlice } from "./querySlice";

export type AppStore = ChatSlice & ChartDrawingsSlice & ConfigSlice & MarketSlice & QuerySlice;

export const useStore = create<AppStore>()((...a) => ({
  ...createChatSlice(...a),
  ...createChartDrawingsSlice(...a),
  ...createConfigSlice(...a),
  ...createMarketSlice(...a),
  ...createQuerySlice(...a),
}));
