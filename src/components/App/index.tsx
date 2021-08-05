import React, { useEffect } from 'react';
import { BrowserRouter as Router, Link, Route } from 'react-router-dom';
import { ProtectedRoute } from '../../routes/ProtectedRoute';

import Home from '@components/Home';
import SignIn from '@components/SignIn';
import SignUp from '@components/SignUp';
import SignOut from '@components/SignOut';
import { SetS3Config } from '../../services';
import Storage from '@aws-amplify/storage';

// import { configureAmplify, SetS3Config } from '../../services';
// import Storage from '@aws-amplify/storage';

const PlaceHolder: React.FC = () => {
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
        console.log(v);
      })
      .catch((e) => {
        console.error(e);
      });
  });

  return <h1>hi</h1>;
};

export const App: React.FC = () => {
  return (
    <div>
      <h2>KnotSet.com</h2>
      <Router>
        <nav>
          <Link to="/">Home</Link> | <Link to="/signup">Sign Up</Link> | <Link to="/signin">Sign In</Link>
          <br />
          <Link to="/dashboard">Dashboard</Link> | <Link to="/signout">Sign Out</Link>
        </nav>
        <hr />

        <Route exact path="/" component={Home} />
        <Route exact path="/signin" component={SignIn} />
        <Route exact path="/signup" component={SignUp} />
        <Route exact path="/signout" component={SignOut} />
        <ProtectedRoute exact path="/dashboard" component={PlaceHolder} />
      </Router>
    </div>
  );
};

export default App;
