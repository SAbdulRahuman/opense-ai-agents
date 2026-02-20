import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Proxy API requests to the Go backend in development
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1"}/:path*`,
      },
    ];
  },

  // Allow images from external sources if needed
  images: {
    remotePatterns: [
      {
        protocol: "https",
        hostname: "**",
      },
    ],
  },

  // Optimize production builds
  output: "standalone",
};

export default nextConfig;
