/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  images: { unoptimized: true },
  trailingSlash: true,
  env: {
    NEXT_PUBLIC_TESTPAY_MODE: process.env.NEXT_PUBLIC_TESTPAY_MODE || 'local',
  },
};
module.exports = nextConfig;
