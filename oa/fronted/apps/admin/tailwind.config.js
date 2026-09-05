/** @type {import('tailwindcss').Config} */
export default {
  content: ["./apps/*/src/**/*.{js,ts,jsx,tsx}", "./packages/ui/src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      colors: {
        primary: {
          50: "#eff6ff",
          100: "#dbeafe",
          200: "#bfdbfe",
          300: "#93c5fd",
          400: "#60a5fa",
          500: "#3b82f6",
          600: "#2563eb",
          700: "#1d4ed8",
          800: "#1e40af",
          900: "#1e3a8a"
        }
      }
    }
  },
  plugins: [],
  corePlugins: {
    preflight: false // 禁用 Tailwind 的基础样式，避免与 Ant Design 冲突
  }
};
