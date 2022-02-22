module.exports = {
  plugins: [
    [
      "@vuepress/google-analytics",
      {
        ga: process.env.GA,
      }
    ],
    [
      "sitemap", {
        "hostname": "https://omegaatt.com/",
        "exclude": ['/404.html'],
        "dateFormatter": time => {
            return new Date(time).toISOString().slice(0,10);
        }
      }
    ]
  ],
  title: "omegaatt",
  description: "",
  dest: "public",
  head: [
    [
      "link",
      {
        "rel": "icon",
        "href": "/assets/favicon.png"
      }
    ],
    [
      "meta",
      {
        "name": "viewport",
        "content": "width=device-width,initial-scale=1,user-scalable=no"
      }
    ]
  ],
  theme: "reco",
  themeConfig: {
    noFoundPageByTencent: false,
    "nav": [
      {
        "text": "Home",
        "link": "/",
        "icon": "reco-home"
      },
      {
        "text": "TimeLine",
        "link": "/timeline/",
        "icon": "reco-date"
      },
      {
        "text": "Photo",
        "icon": "reco-message",
        "items": [
          {
            "text": "攝影專欄",
            "link": "/photo/gallery/"
          }
        ]
      },
      {
        "text": "Contact",
        "icon": "reco-message",
        // https://vuepress-theme-reco.recoluan.com/en/views/1.x/configJs.html#icon
        "items": [
          {
            "text": "GitHub",
            "link": "https://github.com/omegaatt36",
            "icon": "reco-github"
          },
          {
            "text": "Instagram",
            "link": "https://www.instagram.com/chih.hong/",
            "icon": "reco-account"
          },
          {
            "text": "facebook",
            "link": "https://www.facebook.com/Raiven.Kao",
            "icon": "reco-facebook"
          },
          {
            "text": "mail",
            "link": "mailto:omagaatt36@gmail.com",
            "icon": "reco-mail"
          },
        ]
      }
    ],
    sidebar: {
      '/photo/gallery/': [
        '',
        'that'
      ],
    },
    type: "blog",
    blogConfig: {
      "category": {
        "location": 2,
        "text": "Category"
      },
      "tag": {
        "location": 3,
        "text": "Tag"
      }
    },
    friendLink: [
      {
        "title": "instagram",
        "desc": "chih.hong",
        "avatar":"",
        "email": "omegaatt36@gmail.com",
        "link": "https://www.instagram.com/chih.hong/"
      },
      {
        "title": "vuepress-theme-reco",
        "desc": "A simple and beautiful vuepress Blog & Doc theme.",
        "avatar": "https://vuepress-theme-reco.recoluan.com/icon_vuepress_reco.png",
        "link": "https://vuepress-theme-reco.recoluan.com"
      }
    ],
    search: true,
    searchMaxSuggestions: 10,
    lastUpdated: "Last Updated",
    author: "Raiven",
    authorAvatar: "/assets/avatar.jpg",
    startYear: "2020"
  },
  "markdown": {
    "lineNumbers": true
  }
}