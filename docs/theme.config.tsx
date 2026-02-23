import React from 'react'
import { DocsThemeConfig } from 'nextra-theme-docs'

const config: DocsThemeConfig = {
  logo: (
    <span style={{ fontWeight: 800, fontSize: '1.2em' }}>
      Boatman
    </span>
  ),
  project: {
    link: 'https://github.com/philjestin/boatman',
  },
  docsRepositoryBase: 'https://github.com/philjestin/boatman/tree/main/docs',
  footer: {
    text: 'Boatman Ecosystem Documentation',
  },
  head: (
    <>
      <meta name="viewport" content="width=device-width, initial-scale=1.0" />
      <meta property="og:title" content="Boatman Documentation" />
      <meta property="og:description" content="AI-powered autonomous development with BoatmanMode CLI and Boatman Desktop" />
    </>
  ),
  useNextSeoProps() {
    return {
      titleTemplate: '%s - Boatman Docs'
    }
  },
  sidebar: {
    defaultMenuCollapseLevel: 1,
    toggleButton: true,
  },
  toc: {
    backToTop: true,
  },
  navigation: {
    prev: true,
    next: true,
  },
  primaryHue: 210,
  banner: {
    key: 'boatman-docs-v1',
    text: 'Boatman Ecosystem documentation is live!',
  },
}

export default config
