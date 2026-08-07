import prettierConfig from 'eslint-config-prettier'
import vuetify from 'eslint-config-vuetify'

export default vuetify({
  ts: false,
}, {
  rules: {
    'vue/block-lang': 'off',
    'vue/script-indent': 'off',
    'vue/attributes-order': 'off',
    'vue/padding-line-between-tags': 'off',
    'vue/custom-event-name-casing': 'off',
    'unicorn/no-this-outside-of-class': 'off',
    'unicorn/catch-error-name': 'off',
    'unicorn/switch-case-braces': 'off',
    'unicorn/no-useless-switch-case': 'off',
    'perfectionist/sort-named-imports': 'off',
    'perfectionist/sort-imports': 'off',
    'perfectionist/sort-object-types': 'off',
  },
}).then(config => [
  {
    ignores: ['**/typed-router.d.ts', 'dist/**'],
  },
  ...config,
  prettierConfig, // Ensures ESLint leaves formatting to Prettier
])
