import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: ["class"],
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        bg: "#ECEAE5",
        background: "#ECEAE5",
        canvas: "#ECEAE5",
        surface: "rgba(255, 255, 255, 0.62)",
        card: {
          DEFAULT: "rgba(255, 255, 255, 0.62)",
          foreground: "#1E1A17",
        },
        borderSubtle: "rgba(30, 26, 23, 0.08)",
        border: "rgba(30, 26, 23, 0.08)",
        primary: {
          DEFAULT: "#B87C3B",
          foreground: "#FFFFFF",
          hover: "#9c642c",
        },
        primaryHover: "#9c642c",
        textMain: "#1E1A17",
        textMuted: "#4A4540",
        textSubtle: "#7C7770",
        foreground: "#1E1A17",
        muted: {
          DEFAULT: "rgba(30, 26, 23, 0.05)",
          foreground: "#4A4540",
        },
        obsidian: "#151310",
        status: {
          online: "#10B981",
          degraded: "#B87C3B",
          offline: "#EF4444",
        },
        success: "#10B981",
        warning: "#B87C3B",
        danger: "#EF4444",
      },
      fontFamily: {
        sans: ["Inter", "-apple-system", "BlinkMacSystemFont", "Segoe UI", "Roboto", "sans-serif"],
        mono: ["JetBrains Mono", "monospace"],
      },
      boxShadow: {
        glass: "inset 0 1px 1px 0 rgba(255, 255, 255, 0.85), 0 12px 32px -4px rgba(30, 26, 23, 0.04), 0 4px 12px -2px rgba(184, 124, 59, 0.03)",
        card: "inset 0 1px 1px 0 rgba(255, 255, 255, 0.85), 0 12px 32px -4px rgba(30, 26, 23, 0.04), 0 4px 12px -2px rgba(184, 124, 59, 0.03)",
        "glass-glow": "inset 0 1px 1.5px 0 rgba(255, 255, 255, 0.95), 0 16px 36px -4px rgba(30, 26, 23, 0.06), 0 0 28px -2px rgba(184, 124, 59, 0.08)",
      },
    },
  },
  plugins: [],
};

export default config;
