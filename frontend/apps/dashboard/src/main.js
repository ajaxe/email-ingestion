/**
 * main.ts
 *
 * Bootstraps Vuetify and other plugins then mounts the App`
 */

import { createApp } from "vue";
import { registerPlugins } from "@/plugins";
import Main from "./Main.vue";
import router from "./router";

// Styles
import "unfonts.css";
import "virtual:uno.css";
import "./styles/main.scss";

const app = createApp(Main);

registerPlugins(app);

router.isReady().then(() => app.mount("#app"));
