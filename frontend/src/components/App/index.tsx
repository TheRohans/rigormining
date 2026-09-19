import React, { useEffect, useState } from 'react';
import { BrowserRouter, Route, Switch } from 'react-router-dom';
import { ProtectedRoute } from '../../routes/ProtectedRoute';

import Home from '@components/Home';
import SignIn from '@components/SignIn';
import Library from '@components/Library';
import ItemDetail from '@components/Library/ItemDetail';
import Sync from '@components/Sync';
import Settings from '@components/Settings';
import Extension from '@components/Extension';

import Navigation from './Navigation';
import { api, WhoAmI } from '../../api/client';

export const App: React.FC = () => {
  const [user, setUser] = useState<WhoAmI | null>(null);
  const [checked, setChecked] = useState(false);

  useEffect(() => {
    api
      .whoami()
      .then(setUser)
      .catch(() => setUser(null))
      .finally(() => setChecked(true));
  }, []);

  if (!checked) {
    return null;
  }

  return (
    <BrowserRouter>
      <Switch>
        <Route exact path="/">
          <Navigation loggedIn={!!user} />
          <Home />
        </Route>

        <Route exact path="/signin">
          <SignIn />
        </Route>

        <ProtectedRoute exact path="/library" component={Library} loggedIn={!!user} />
        <ProtectedRoute exact path="/library/:id" component={ItemDetail} loggedIn={!!user} />
        <ProtectedRoute exact path="/sync" component={Sync} loggedIn={!!user} />
        <ProtectedRoute exact path="/settings" component={Settings} loggedIn={!!user} />
        <ProtectedRoute exact path="/extension" component={Extension} loggedIn={!!user} />
      </Switch>
    </BrowserRouter>
  );
};

export default App;
