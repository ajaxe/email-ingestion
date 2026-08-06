import prettierConfig from 'eslint-config-prettier'
import vuetify from 'eslint-config-vuetify'

export default vuetify({
  ts: false,
}, {
  rules: {
    'vue/block-lang': 'off',
    'vue/script-indent': 'off',
    'vue/attributes-order': 'off',
    'perfectionist/sort-named-imports': 'off',
    'perfectionist/sort-imports': 'off',
    'perfectionist/sort-object-types': 'off',
  },
}).then(config => [
  ...config,
  prettierConfig, // Ensures ESLint leaves formatting to Prettier
])
