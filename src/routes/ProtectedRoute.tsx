import React from 'react';
import { Redirect, Route } from 'react-router-dom';
import { Auth } from 'aws-amplify';

export const useAuth = () => {
  let isAuthed = false;

  try {
    isAuthed = JSON.parse(localStorage.getItem('isAuthenticated')) ?? false;
  } catch (e) {
    // can't parse local storage
    console.error('ls', e);
  }

  if (!isAuthed) {
    // we don't have a local state, or the state is false
    // check if we are authed for real
    Auth.currentAuthenticatedUser()
      .then((user) => {
        if (user) {
          isAuthed = true;
          localStorage.setItem('isAuthenticated', JSON.stringify(isAuthed));
        }
      })
      .catch((e) => {
        localStorage.removeItem('isAuthenticated');
        // console.error('er', e);
      });
  }

  return isAuthed;
};

export const ProtectedRoute = (props: any) => {
  const loggedIn = useAuth();

  return (
    <Route
      path={props.path}
      render={(data) =>
        loggedIn ? <props.component {...data}></props.component> : <Redirect to={{ pathname: '/signin' }}></Redirect>
      }
    ></Route>
  );
};

export default ProtectedRoute;
