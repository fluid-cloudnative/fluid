import routes from './routes';
import locales from './locales';

const menus = [
  {
    parent: 'topbar',
    name: 'fluid',
    title: 'fluid',
    icon: 'cluster',
    order: 0,
    desc: 'Fluid, elastic data abstraction and acceleration for BigData&#x2F;AI applications in cloud.',
    skipAuth: true,
  },
];

const extensionConfig = {
  routes,
  menus,
  locales,
};

export default extensionConfig;
