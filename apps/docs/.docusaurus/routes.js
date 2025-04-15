import React from 'react';
import ComponentCreator from '@docusaurus/ComponentCreator';

export default [
  {
    path: '/expo-open-ota/__docusaurus/debug',
    component: ComponentCreator('/expo-open-ota/__docusaurus/debug', '3e2'),
    exact: true
  },
  {
    path: '/expo-open-ota/__docusaurus/debug/config',
    component: ComponentCreator('/expo-open-ota/__docusaurus/debug/config', '9e3'),
    exact: true
  },
  {
    path: '/expo-open-ota/__docusaurus/debug/content',
    component: ComponentCreator('/expo-open-ota/__docusaurus/debug/content', '6c9'),
    exact: true
  },
  {
    path: '/expo-open-ota/__docusaurus/debug/globalData',
    component: ComponentCreator('/expo-open-ota/__docusaurus/debug/globalData', 'd0d'),
    exact: true
  },
  {
    path: '/expo-open-ota/__docusaurus/debug/metadata',
    component: ComponentCreator('/expo-open-ota/__docusaurus/debug/metadata', 'a1e'),
    exact: true
  },
  {
    path: '/expo-open-ota/__docusaurus/debug/registry',
    component: ComponentCreator('/expo-open-ota/__docusaurus/debug/registry', 'd08'),
    exact: true
  },
  {
    path: '/expo-open-ota/__docusaurus/debug/routes',
    component: ComponentCreator('/expo-open-ota/__docusaurus/debug/routes', 'fda'),
    exact: true
  },
  {
    path: '/expo-open-ota/markdown-page',
    component: ComponentCreator('/expo-open-ota/markdown-page', '23e'),
    exact: true
  },
  {
    path: '/expo-open-ota/docs',
    component: ComponentCreator('/expo-open-ota/docs', 'dd6'),
    routes: [
      {
        path: '/expo-open-ota/docs',
        component: ComponentCreator('/expo-open-ota/docs', 'a89'),
        routes: [
          {
            path: '/expo-open-ota/docs',
            component: ComponentCreator('/expo-open-ota/docs', 'a16'),
            routes: [
              {
                path: '/expo-open-ota/docs/advanced/prometheus',
                component: ComponentCreator('/expo-open-ota/docs/advanced/prometheus', '08c'),
                exact: true,
                sidebar: "docSidebar"
              },
              {
                path: '/expo-open-ota/docs/cache',
                component: ComponentCreator('/expo-open-ota/docs/cache', '624'),
                exact: true,
                sidebar: "docSidebar"
              },
              {
                path: '/expo-open-ota/docs/category/deployment',
                component: ComponentCreator('/expo-open-ota/docs/category/deployment', '62f'),
                exact: true,
                sidebar: "docSidebar"
              },
              {
                path: '/expo-open-ota/docs/cdn/cloudfront',
                component: ComponentCreator('/expo-open-ota/docs/cdn/cloudfront', 'dd9'),
                exact: true,
                sidebar: "docSidebar"
              },
              {
                path: '/expo-open-ota/docs/cdn/intro',
                component: ComponentCreator('/expo-open-ota/docs/cdn/intro', '6ea'),
                exact: true,
                sidebar: "docSidebar"
              },
              {
                path: '/expo-open-ota/docs/dashboard',
                component: ComponentCreator('/expo-open-ota/docs/dashboard', '617'),
                exact: true,
                sidebar: "docSidebar"
              },
              {
                path: '/expo-open-ota/docs/deployment/custom',
                component: ComponentCreator('/expo-open-ota/docs/deployment/custom', '588'),
                exact: true,
                sidebar: "docSidebar"
              },
              {
                path: '/expo-open-ota/docs/deployment/helm',
                component: ComponentCreator('/expo-open-ota/docs/deployment/helm', 'e7c'),
                exact: true,
                sidebar: "docSidebar"
              },
              {
                path: '/expo-open-ota/docs/deployment/railway',
                component: ComponentCreator('/expo-open-ota/docs/deployment/railway', '1b9'),
                exact: true,
                sidebar: "docSidebar"
              },
              {
                path: '/expo-open-ota/docs/deployment/testing',
                component: ComponentCreator('/expo-open-ota/docs/deployment/testing', '952'),
                exact: true,
                sidebar: "docSidebar"
              },
              {
                path: '/expo-open-ota/docs/environment',
                component: ComponentCreator('/expo-open-ota/docs/environment', 'afb'),
                exact: true,
                sidebar: "docSidebar"
              },
              {
                path: '/expo-open-ota/docs/eoas/configure',
                component: ComponentCreator('/expo-open-ota/docs/eoas/configure', '523'),
                exact: true,
                sidebar: "docSidebar"
              },
              {
                path: '/expo-open-ota/docs/eoas/intro',
                component: ComponentCreator('/expo-open-ota/docs/eoas/intro', '904'),
                exact: true,
                sidebar: "docSidebar"
              },
              {
                path: '/expo-open-ota/docs/eoas/publish',
                component: ComponentCreator('/expo-open-ota/docs/eoas/publish', 'ece'),
                exact: true,
                sidebar: "docSidebar"
              },
              {
                path: '/expo-open-ota/docs/introduction',
                component: ComponentCreator('/expo-open-ota/docs/introduction', '955'),
                exact: true,
                sidebar: "docSidebar"
              },
              {
                path: '/expo-open-ota/docs/key-store',
                component: ComponentCreator('/expo-open-ota/docs/key-store', '24b'),
                exact: true,
                sidebar: "docSidebar"
              },
              {
                path: '/expo-open-ota/docs/prerequisites',
                component: ComponentCreator('/expo-open-ota/docs/prerequisites', 'edb'),
                exact: true,
                sidebar: "docSidebar"
              },
              {
                path: '/expo-open-ota/docs/storage',
                component: ComponentCreator('/expo-open-ota/docs/storage', '2fb'),
                exact: true,
                sidebar: "docSidebar"
              }
            ]
          }
        ]
      }
    ]
  },
  {
    path: '/expo-open-ota/',
    component: ComponentCreator('/expo-open-ota/', 'ab2'),
    exact: true
  },
  {
    path: '*',
    component: ComponentCreator('*'),
  },
];
