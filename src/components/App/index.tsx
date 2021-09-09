import React, { useState } from 'react';
import { BrowserRouter, Route, Switch } from 'react-router-dom';
import { Auth } from 'aws-amplify';
import { ProtectedRoute } from '../../routes/ProtectedRoute';

import Home from '@components/Home';
import SignIn from '@components/SignIn';
import SignUp from '@components/SignUp';
import SignOut from '@components/SignOut';
import Library from '@components/Library';
import Highlights from '@components/Highlights';
import Import from '@components/Highlights/import';

import Navigation from './Navigation';

export const App: React.FC = () => {
  const [loggedIn, setLoggedIn] = useState(false);

  const logIn = (email: string, password: string): Promise<any> => {
    console.log('submitting');

    return Auth.signIn({
      username: email,
      password,
    }).then((user) => {
      setLoggedIn(true);
      return user;
    });
  };

  const logOut = (): Promise<void> => {
    return Auth.signOut().then(() => {
      setLoggedIn(false);
      return;
    });
  };

  return (
    <BrowserRouter>
      <Switch>
        <Route exact path="/">
          <Navigation loggedIn={loggedIn} />
          <Home />
        </Route>

        <Route exact path="/signin">
          <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
            {/* <Navigation loggedIn={loggedIn} /> */}
            <div className="py-10">
              <main>
                <SignIn logIn={logIn} />
              </main>
            </div>
          </div>
        </Route>

        <Route exact path="/signout">
          {/* <Navigation loggedIn={loggedIn} /> */}
          <SignOut logOut={logOut} />
        </Route>

        <Route exact path="/signup" component={SignUp} />

        <ProtectedRoute exact loggedIn={loggedIn} path="/library" component={Library} />
        <ProtectedRoute exact loggedIn={loggedIn} path="/highlights" component={Highlights} />
        <ProtectedRoute exact loggedIn={loggedIn} path="/import" component={Import} />
      </Switch>
    </BrowserRouter>
  );
};

export default App;
