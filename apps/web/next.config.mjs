/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  async headers() {
    return [
      {
        source: "/(.*)",
        headers: [
          {
            key: "Content-Security-Policy",
            // strict lock: only allow fetch/xhr to itself, the backend API, and Razorpay
            value: "connect-src 'self' https://halomailapi.zeroaxiis.tech http://localhost:8080 http://127.0.0.1:8080 https://api.razorpay.com https://*.razorpay.com;",
          },
        ],
      },
    ];
  },
};

export default nextConfig;
