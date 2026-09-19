import React from 'react';
import { Redirect, Route, RouteProps } from 'react-router-dom';

type ProtectedRouteProps = RouteProps & {
  component: React.ComponentType<any>;
  loggedIn: boolean;
};

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ component: Component, loggedIn, ...rest }) => (
  <Route
    {...rest}
    render={(routeProps) => (loggedIn ? <Component {...routeProps} /> : <Redirect to={{ pathname: '/signin' }} />)}
  />
);

export default ProtectedRoute;
