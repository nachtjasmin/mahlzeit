const colors = require("tailwindcss/colors");

const config = {
  content: ["./templates/**/*.tmpl"],
  theme: {
    extend: {
      colors: {
        accent: colors.green,
        neutral: colors.stone,
      },
    },
  },
  plugins: [require("tailwindcss/nesting"), require("@tailwindcss/forms")],
};

module.exports = config;
