import React, { useEffect, useState } from 'react';
import { BrowserRouter as Router, Link, Route, Switch } from 'react-router-dom';
import { ProtectedRoute, useAuth } from '../../routes/ProtectedRoute';

import Home from '@components/Home';
import SignIn from '@components/SignIn';
import SignUp from '@components/SignUp';
import SignOut from '@components/SignOut';
import { SetS3Config } from '../../services';
import Storage from '@aws-amplify/storage';
import Navigation from './Navigation';

// import { configureAmplify, SetS3Config } from '../../services';
// import Storage from '@aws-amplify/storage';

const PlaceHolder: React.FC = () => {
  const [booklist, setBookList] = useState([]);

  useEffect(() => {
    SetS3Config('private.knotset.com', 'public');

    // Storage.put('profile.json', JSON.stringify({ wtf: 'mate' }), {
    //   level: 'private',
    //   contentType: 'application/json',
    // })
    //   .then((result) => console.log(result))
    //   .catch((err) => console.log(err));

    Storage.get('library/rohan/metadata.json', {
      level: 'public',
      download: true,
      contentType: 'application/json',
    })
      .then((v) => (v as any)?.Body?.text())
      .then((v) => {
        const json = JSON.parse(v);
        setBookList(json.children);
      })
      .catch((e) => {
        console.error(e);
      });
  }, []);

  const downloadBook = (name: string) => {
    Storage.get(`library/rohan/${name}`, {
      level: 'public',
    })
      .then((v) => {
        window.location.href = v as string;
      })
      .catch((e) => {
        console.error(e);
      });
  };

  return (
    <div>
      {booklist.map((b: any) => (
        <div key={b.name}>
          <a onClick={() => downloadBook(b?.name)}>{b?.name}</a>
        </div>
      ))}
    </div>
  );
};

export const App: React.FC = () => {
  const loggedIn = useAuth();

  return (
    <>
      <Router>
        {/* <Nav authed={isAuthenticated()} /> */}
        <Navigation loggedIn={loggedIn} />
        <Switch>
          <Route exact path="/" component={Home} />
          <Route exact path="/signin" component={SignIn} />
          <Route exact path="/signup" component={SignUp} />
          <Route exact path="/signout" component={SignOut} />
          <ProtectedRoute exact path="/dashboard" component={PlaceHolder} />
        </Switch>
      </Router>
    </>
  );
};

export default App;
