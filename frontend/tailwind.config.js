module.exports = {
	future: {
		removeDeprecatedGapUtilities: true,
		purgeLayersByDefault: true,
	},
	purge: {
		enabled: true,
		mode: 'all',
		content: ['./src/*.svelte'],
		options: {
			safelist: ['border-yellow-800', 'border-green-900', 'line-through', 'bg-gray-700', 'bg-gray-800'],
		}
	},
	theme: {
		extend: {},
	},
	variants: {},
	plugins: [],
}
