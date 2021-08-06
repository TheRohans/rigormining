import React from 'react';
import { Redirect, Route } from 'react-router-dom';

export const ProtectedRoute = (props: any) => {
  return (
    <Route
      path={props.path}
      render={(data) =>
        props.loggedIn ? (
          <props.component {...data}></props.component>
        ) : (
          <Redirect to={{ pathname: '/signin' }}></Redirect>
        )
      }
    ></Route>
  );
};

export default ProtectedRoute;
