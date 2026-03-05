/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{html,ts}",
  ],
  theme: {
    extend: {
      colors: {
        'aurora-base': '#e1c78c',
        'aurora-primary': '#eda011',
        'aurora-secondary': '#db6516',
        'aurora-dark': '#7a6949',
        'aurora-light': '#adad8e',
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
