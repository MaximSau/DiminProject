/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        darkBg: '#08080a',
        darkCard: '#121216',
        darkCardHover: '#17171d',
        darkBorder: 'rgba(255, 255, 255, 0.10)',
        darkSurface: '#101014',
        lightBg: '#f8fafc',
        lightCard: '#ffffff',
        lightBorder: '#e2e8f0',
        lightSurface: '#f1f5f9',
        pulseGreen: '#22c55e',
        pulseBlue: '#38bdf8',
        pulseAmber: '#f59e0b',
        pulseRed: '#ef4444',
      },
      fontFamily: {
        sans: ['Inter', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
      },
    },
  },
  plugins: [],
};
