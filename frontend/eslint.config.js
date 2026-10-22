import js from '@eslint/js'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import globals from 'globals'
import tseslint from 'typescript-eslint'

export default tseslint.config(
  { ignores: ['dist', 'src/shared/api/generated', 'src/app/routeTree.gen.ts'] },
  {
    files: ['**/*.{ts,tsx}'],
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    languageOptions: { ecmaVersion: 2023, globals: globals.browser },
    plugins: { 'react-hooks': reactHooks, 'react-refresh': reactRefresh },
    rules: {
      ...reactHooks.configs.recommended.rules,
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
      // Features talk to the API through shared/api and to Firebase & co. through shared/platform.
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            {
              group: ['firebase', 'firebase/*'],
              message: 'Use the adapters in @/shared/platform instead.',
            },
            {
              group: ['@/features/*/*'],
              message: 'Import a feature through its public index, not its internals.',
            },
          ],
        },
      ],
    },
  },
  {
    // Route files export a Route object; the router plugin handles code splitting and HMR.
    files: ['src/app/routes/**'],
    rules: { 'react-refresh/only-export-components': 'off' },
  },
  {
    // Adapters are the only place allowed to touch third-party platform SDKs.
    files: ['src/shared/platform/**'],
    rules: { 'no-restricted-imports': 'off' },
  },
)
