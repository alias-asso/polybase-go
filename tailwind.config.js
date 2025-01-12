export default {
  content: ["./views/**/*.templ"],
  darkMode: 'selector',
  theme: {
    extend: {
      colors: {
        base: {
          100: 'var(--base-100)',
          200: 'var(--base-200)',
          300: 'var(--base-300)',
          400: 'var(--base-400)',
          500: 'var(--base-500)',
          600: 'var(--base-600)',
          700: 'var(--base-700)',
          800: 'var(--base-800)',
          900: 'var(--base-900)',
        },
        accent: {
          100: '#d1def9',
          200: '#7a94cd',
          300: '#4968b1',
          400: '#2648a1',
          500: '#102d8c',
          600: '#061d7c',
          700: '#020e68',
          800: '#020e51',
          900: '#030f37',
        }
      }
    }
  }
};
