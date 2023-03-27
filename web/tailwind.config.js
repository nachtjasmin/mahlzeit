const colors = require("tailwindcss/colors");

const config = {
	content: ["./templates/**/*.tmpl"],
	theme: {
		extend: {
			fontFamily: {
				// taken from: https://modernfontstacks.com/
				"old-style": [
					"Iowan Old Style",
					"Palatino Linotype",
					"URW Palladio L",
					"P052",
					"serif",
				],
				"geometric-humanist": [
					"Avenir",
					"Avenir Next LT Pro",
					"Montserrat",
					"Corbel",
					"URW Gothic",
					"source-sans-pro",
					"sans-serif",
				],
			},
			colors: {
				// Palette taken from Kate: https://iamkate.com/data/12-bit-rainbow/
				rainbow: {
					purple: "#817",
					darkred: "#a35",
					red: "#c66",
					orange: "#e94",
					yellow: "#ed0",
					green: "#9d5",
					lightgreen: "#4d8",
					lightcyan: "#2cb",
					cyan: "#0bc",
					blue: "#09c",
					darkblue: "#36b",
					darkpurple: "#639",
				},
				accent: colors.green,
				neutral: colors.stone,
			},
		},
	},
	plugins: [require("tailwindcss/nesting"), require("@tailwindcss/forms")],
};

module.exports = config;
