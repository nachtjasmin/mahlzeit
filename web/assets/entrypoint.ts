import "vite/modulepreload-polyfill";
import "htmx.org";

// Import the necessary tailwind stuff first.
import "tailwindcss/base.css";
import "tailwindcss/components.css";

// After that, load the custom CSS.
import "./css/_index.pcss";

// Finally, load the utilities. This helps us, because now the utilities have a higher
// priority than our custom CSS, so we can override certain things easily.
import "tailwindcss/utilities.css";
