const colors = require('tailwindcss/colors');

module.exports = {
  purge: {
    enabled: true,
    content: ['src/**/*.html', 'src/**/*.tsx', 'src/**/*.jsx', 'src/**/*.js', 'src/**/*.ts'],
  },
  darkMode: 'media', // or 'media' or 'class'
  theme: {
    colors: {
      transparent: 'transparent',
      current: 'currentColor',

      black: colors.black,
      white: colors.white,
      gray: colors.coolGray,
      red: colors.red,
      yellow: colors.amber,
      green: colors.emerald,
      blue: colors.blue,
      indigo: colors.indigo,
      purple: colors.violet,
      pink: colors.pink,
    },
  },
  variants: {
    extend: {},
  },
  plugins: [],
};
