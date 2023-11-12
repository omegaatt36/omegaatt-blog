import { defineUserConfig } from 'vuepress'
import { sitemapPlugin } from "vuepress-plugin-sitemap2";
import { googleAnalyticsPlugin } from '@vuepress/plugin-google-analytics'
import recoTheme from 'vuepress-theme-reco'

export default defineUserConfig({
  plugins: [
    sitemapPlugin({
      hostname: process.env.HOST_NAME || ''
    }),
    googleAnalyticsPlugin({
      id: process.env.GA || '',
    }),
  ],
  title: 'omegaatt',
  description: `Raiven Kao's Blog`,
  theme: recoTheme({
    style: '@vuepress-reco/style-default',
    logo: '/favicon.ico',
    author: 'Raiven Kao',
    authorAvatar: '/assets/avatar.webp',
    startYear: 2020,
    docsRepo: 'https://github.com/omegaatt36/omegaatt36-blog',
    docsBranch: 'main',
    // docsDir: 'example',
    lastUpdatedText: '',
    // series: {
    //   '/blogs/develop/': [
    //     {
    //       text: 'module one',
    //       children: ['home', 'theme']
    //     },
    //     {
    //       text: 'module two',
    //       children: ['api', 'plugin']
    //     }
    //   ]
    // },
    autoSetBlogCategories: true,
    autoAddCategoryToNavbar: true,
    navbar:
    [
      {
        text: 'Contact',
        children: [
          {
            "text": "mail",
            "link": "mailto:omagaatt36@gmail.com",
            "icon": "Email"
          },
          {
            "text": "GitHub",
            "link": "https://github.com/omegaatt36",
            "icon": "LogoGithub"
          },
          {
            "text": "Instagram",
            "link": "https://www.instagram.com/chih.hong/",
            "icon": "LogoInstagram"
          },
          {
            "text": "facebook",
            "link": "https://www.facebook.com/Raiven.Kao",
            "icon": "LogoGithub",
          },
          {
            "text": "linkedin",
            "link": "https://www.linkedin.com/in/raiven/",
            "icon": "LogoLinkedin"
          },
        ]
      },
    ],
    // bulletin: {
    //   body: [
    //     {
    //       type: 'buttongroup',
    //       children: [
    //         {
    //           text: 'Instagram',
    //           link: 'https://www.instagram.com/chih.hong/'
    //         }
    //       ]
    //     }
    //   ],
    // },
  }),
  debug: true,
})
