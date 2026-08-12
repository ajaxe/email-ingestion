import UnoCSS from "unocss/vite";
import VueRouter from "vue-router/vite";
import { fileURLToPath, URL } from "node:url";
import Vue from "@vitejs/plugin-vue";
import Fonts from "unplugin-fonts/vite";
import { defineConfig, loadEnv } from "vite";
import Vuetify, { transformAssetUrls } from "vite-plugin-vuetify";
import * as fs from "node:fs";

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");

  console.log(`mode: ${mode}, env.VITE_DEV_KEY_FILE: ${env.VITE_DEV_KEY_FILE}`)

  return {
    plugins: [
      VueRouter({ dts: "src/typed-router.d.ts" }),
      // https://github.com/vuetifyjs/vuetify-loader/tree/master/packages/vite-plugin#readme
      Vue({
        template: { transformAssetUrls },
      }),
      Vuetify({
        autoImport: true,
        styles: {
          configFile: "src/styles/settings.scss",
        },
      }),
      Fonts({
        fontsource: {
          families: [
            {
              name: "Roboto Mono",
              weights: [400, 700],
            },
            {
              name: "Roboto",
              weights: [100, 300, 400, 500, 700, 900],
              styles: ["normal", "italic"],
            },
          ],
        },
      }),
      UnoCSS(),
    ],
    define: { "process.env": {} },
    resolve: {
      alias: {
        "@": fileURLToPath(new URL("src", import.meta.url)),
      },
      extensions: [".js", ".json", ".jsx", ".mjs", ".ts", ".tsx", ".vue"],
    },
    server: {
      port: 3000,
      https: env.VITE_DEV_KEY_FILE
        ? {
            key: fs.readFileSync(env.VITE_DEV_KEY_FILE),
            cert: fs.readFileSync(env.VITE_DEV_CERT_FILE),
          }
        : null,
      proxy: {
        "/app": {
          target: "http://localhost:8080",
          changeOrigin: true,
        },
      },
    },
  };
});
