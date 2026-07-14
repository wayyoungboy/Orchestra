module.exports = {
  root: true,
  env: {
    browser: true,
    es2022: true,
    node: true,
  },
  extends: [
    'eslint:recommended',
    'plugin:vue/vue3-recommended',
    'plugin:@typescript-eslint/recommended',
  ],
  parser: 'vue-eslint-parser',
  parserOptions: {
    ecmaVersion: 'latest',
    parser: '@typescript-eslint/parser',
    sourceType: 'module',
  },
  ignorePatterns: ['dist', 'node_modules', 'coverage'],
  rules: {
    '@typescript-eslint/no-explicit-any': 'off',
    '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_', varsIgnorePattern: '^_' }],
    'vue/multi-word-component-names': 'off',
    // Formatting is enforced by the formatter; lint should focus on correctness.
    'vue/attributes-order': 'off',
    'vue/html-self-closing': 'off',
    'vue/html-closing-bracket-newline': 'off',
    'vue/html-closing-bracket-spacing': 'off',
    'vue/html-indent': 'off',
    'vue/max-attributes-per-line': 'off',
    'vue/first-attribute-linebreak': 'off',
    'vue/require-default-prop': 'off',
    'vue/singleline-html-element-content-newline': 'off',
  },
  overrides: [
    {
      files: ['e2e/**/*.ts'],
      rules: {
        '@typescript-eslint/no-unused-vars': 'off',
      },
    },
    {
      files: ['src/features/chat/components/MessagesList.vue'],
      rules: {
        // This component uses the strict allowlist in renderMarkdownSafe.
        'vue/no-v-html': 'off',
      },
    },
  ],
}
