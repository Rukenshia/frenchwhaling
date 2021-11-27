module.exports = {
  purge: {
    enabled: true,
    mode: 'all',
    content: ['./src/*.svelte'],
    options: {
      safelist: [
        'border-yellow-800',
        'border-green-900',
        'line-through',
        'bg-gray-700',
        'bg-gray-800',
        'text-yellow-600',
        'group-hover',
        'group',
      ],
    },
  },
  theme: {},
  variants: {
    extend: {
      display: ['group-hover'],
    },
  },
  plugins: [],
};
