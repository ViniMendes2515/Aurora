/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{html,ts}",
  ],
  theme: {
    extend: {
      colors: {
        'aurora-base': 'rgb(2 24 38)',
        'aurora-primary': 'rgb(15 140 140)',
        'aurora-secondary': 'rgb(2 89 89)',
        'aurora-dark': 'rgb(1 40 64)',
        'aurora-light': 'rgb(1 28 38)',
        'aurora-accent': 'rgb(15 140 140)',
        'aurora-deep': 'rgb(1 18 38)',
        'aurora-mid': 'rgb(2 24 38)',
        'aurora-surface': 'rgb(10 37 53)',
        'aurora-border': 'rgb(13 51 72)',
        'aurora-text': 'rgb(168 197 216)',
        'aurora-text-muted': 'rgb(90 128 153)',
      },
      fontFamily: {
        display: ['"Yeseva One"', 'serif'],
        sans: ['"Josefin Sans"', 'system-ui', 'sans-serif'],
      },
      backgroundImage: {
        'aurora-gradient': 'linear-gradient(135deg, #012840 0%, #011C26 50%, #021826 100%)',
        'aurora-glow': 'radial-gradient(ellipse at top, #025959 0%, transparent 60%)',
        'teal-glow': 'radial-gradient(circle at 50% 0%, rgba(15,140,140,0.15) 0%, transparent 70%)',
      },
      boxShadow: {
        'aurora': '0 4px 24px rgba(15,140,140,0.15)',
        'aurora-lg': '0 8px 48px rgba(15,140,140,0.2)',
        'aurora-sm': '0 2px 12px rgba(1,40,64,0.4)',
        'inner-aurora': 'inset 0 1px 0 rgba(15,140,140,0.1)',
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'glow': 'glow 2s ease-in-out infinite alternate',
        'float': 'float 6s ease-in-out infinite',
        'fade-in': 'fadeIn 0.6s ease forwards',
        'slide-up': 'slideUp 0.5s ease forwards',
      },
      keyframes: {
        glow: {
          '0%': { boxShadow: '0 0 5px rgba(15,140,140,0.3)' },
          '100%': { boxShadow: '0 0 20px rgba(15,140,140,0.6), 0 0 40px rgba(15,140,140,0.2)' },
        },
        float: {
          '0%, 100%': { transform: 'translateY(0)' },
          '50%': { transform: 'translateY(-6px)' },
        },
        fadeIn: {
          from: { opacity: '0' },
          to: { opacity: '1' },
        },
        slideUp: {
          from: { opacity: '0', transform: 'translateY(12px)' },
          to: { opacity: '1', transform: 'translateY(0)' },
        },
      },
    },
  },
  safelist: [
    'ml-64', 'ml-20', 'ml-0',
    'w-64', 'w-20',
    'px-3.5', 'px-0', 'px-3', 'px-2',
    'p-3', 'py-3',
    'justify-center',
  ],
  plugins: [],
}
