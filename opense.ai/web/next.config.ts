import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Static export for embedding into Go binary
  output: "export",

  // Required for static export — no server-side image optimization
  images: {
    unoptimized: true,
  },

  // Trailing slash ensures each route gets its own directory with index.html
  trailingSlash: true,
};

export default nextConfig;
