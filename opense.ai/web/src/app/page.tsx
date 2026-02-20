import { MarketOverview, Watchlist, FIIDIIBar, TopMovers } from "@/components/dashboard";

export const metadata = {
  title: "Dashboard | OpeNSE.ai",
};

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Indian stock market overview — NSE live data
        </p>
      </div>

      {/* Market Indices */}
      <MarketOverview />

      {/* Main grid: Watchlist + FII/DII + Top Movers */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-1">
          <Watchlist />
        </div>
        <div className="lg:col-span-1">
          <FIIDIIBar />
        </div>
        <div className="lg:col-span-1">
          <TopMovers />
        </div>
      </div>
    </div>
  );
}
