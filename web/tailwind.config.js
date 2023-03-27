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
				accent: colors.green,
				neutral: colors.stone,
			},
		},
	},
	plugins: [require("tailwindcss/nesting"), require("@tailwindcss/forms")],
};

module.exports = config;
