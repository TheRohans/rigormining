import React from 'react';
import { Redirect, Route } from 'react-router-dom';

export const ProtectedRoute = (props: any) => {
  const loggedIn = !!localStorage.getItem('isAuthenticated');

  return (
    <Route
      path={props.path}
      render={(data) =>
        loggedIn ? (
          <props.component {...props} loggedIn={loggedIn} {...data}></props.component>
        ) : (
          <Redirect to={{ pathname: '/signin' }}></Redirect>
        )
      }
    ></Route>
  );
};

export default ProtectedRoute;
