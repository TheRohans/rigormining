import Hello from '@components/Home/index';
import World from '@components/Page2/index';

export const routes = [
  {
    component: Hello,
    exact: true,
    path: '/',
  },
  {
    component: World,
    exact: false,
    path: '/page2',
  },
];
