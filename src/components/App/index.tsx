import React from 'react';
import { BrowserRouter as Router, Link, Route } from 'react-router-dom';
import { ProtectedRoute } from '../../routes/ProtectedRoute';

import Home from '@components/Home';
import SignIn from '@components/SignIn';
import SignUp from '@components/SignUp';
import SignOut from '@components/SignOut';

// import { configureAmplify, SetS3Config } from '../../services';
// import Storage from '@aws-amplify/storage';

const PlaceHolder: React.FC = () => {
  return <h1>hi</h1>;
};

export const App: React.FC = () => {
  // const uploadImage = () => {
  //   SetS3Config('my-test-bucket-amplify', 'protected');

  //   // Storage.put(`userimages/${upload.files[0].name}`, upload.files[0], {
  //   //   contentType: upload.files[0].type,
  //   // })
  //   //   .then((result) => {
  //   //     upload = null;
  //   //     setState({ response: 'Success uploading file!' });
  //   //   })
  //   //   .catch((err) => {
  //   //     setState({ response: `Cannot uploading file: ${err}` });
  //   //   });
  // };

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
