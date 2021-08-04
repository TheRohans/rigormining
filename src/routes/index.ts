import Home from '@components/Home';
import Login from '@components/Login';

export const routes = [
  {
    component: Home,
    exact: true,
    path: '/',
  },
  {
    component: Login,
    exact: false,
    path: '/login',
  },
];
