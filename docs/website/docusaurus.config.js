// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer').themes.github;
const darkCodeTheme = require('prism-react-renderer').themes.dracula;

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'Aruba Cloud CLI',
  tagline: 'Command-line interface for Aruba Cloud APIs',
  favicon: 'img/favicon.ico',

  // Set the production url of your site here
  url: 'https://arubacloud.github.io',
  // Set the /<baseUrl>/ path of your site
  // For GitHub pages, it's often '/<projectName>/'
  baseUrl: '/acloud-cli/',

  // GitHub pages deployment config
  organizationName: 'Arubacloud',
  projectName: 'acloud-cli',
  trailingSlash: false,

  onBrokenLinks: 'throw',
  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },

  // Even if you don't use internalization, you can use this field to set useful
  // metadata like html lang. For example, if your site is Chinese, you may want
  // to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en', 'it'],
    localeConfigs: {
      en: {
        label: 'English',
        direction: 'ltr',
        htmlLang: 'en-US',
        calendar: 'gregory',
      },
      it: {
        label: 'Italiano',
        direction: 'ltr',
        htmlLang: 'it-IT',
        calendar: 'gregory',
      },
    },
  },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          // Enable versioning (disabled during PR checks via DISABLE_VERSIONING env var)
          ...(process.env.DISABLE_VERSIONING !== 'true' && {
            versions: {
              current: {
                label: 'Next',
                path: 'next',
              },
            },
          }),
          // Show last update time
          showLastUpdateTime: true,
          showLastUpdateAuthor: true,
        },
        blog: false,
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      }),
    ],
  ],

  plugins: [
    [
      require.resolve('@easyops-cn/docusaurus-search-local'),
      {
        // Options for the search plugin
        hashed: true,
        language: ['en', 'it'],
        highlightSearchTermsOnTargetPage: true,
        explicitSearchResultPath: true,
        // Fix for GitHub Pages deployment
        indexBlog: false,
        indexPages: false,
        docsRouteBasePath: '/',
        // Ensure search works with baseUrl
        removeDefaultStopWordFilter: false,
        removeDefaultStemmer: false,
      },
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      // Replace with your project's social card
      image: 'img/docusaurus-social-card.jpg',
      navbar: {
        title: 'Aruba Cloud CLI',
        logo: {
          alt: 'Aruba Cloud CLI Logo',
          src: 'img/logo-cloud.png',
          srcDark: 'img/logo-cloud.png',
          width: 32,
          height: 32,
        },
        items: [
          {
            type: 'docSidebar',
            sidebarId: 'tutorialSidebar',
            position: 'left',
            label: 'Documentation',
          },
          {
            type: 'localeDropdown',
            position: 'right',
          },
          {
            type: 'docsVersionDropdown',
            position: 'right',
          },
          {
            href: 'https://github.com/Arubacloud/acloud-cli',
            label: 'GitHub',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Documentation',
            items: [
              {
                label: 'Installation',
                to: '/docs/installation',
              },
              {
                label: 'Resources',
                to: '/docs/resources',
              },
            ],
          },
          {
            title: 'Community',
            items: [
              {
                label: 'GitHub',
                href: 'https://github.com/Arubacloud/acloud-cli',
              },
              {
                label: 'Issues',
                href: 'https://github.com/Arubacloud/acloud-cli/issues',
              },
            ],
          },
          {
            title: 'More',
            items: [
              {
                label: 'Aruba Cloud',
                href: 'https://www.arubacloud.com',
              },
              {
                label: 'Changelog',
                href: 'https://github.com/Arubacloud/acloud-cli/releases',
              },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} Aruba S.p.A. - via San Clemente, 53 - 24036 Ponte San Pietro (BG) P.IVA 01573850516 - C.F. 04552920482 - C.S. € 4.000.000,00 i.v. - Numero REA: BG – 434483 - All rights reserved`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
        additionalLanguages: ['bash', 'powershell', 'yaml'],
      },
    }),
};

module.exports = config;

