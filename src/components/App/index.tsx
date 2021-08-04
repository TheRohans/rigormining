import React from 'react';
import { BrowserRouter as Router, Link, Route } from 'react-router-dom';

import { routes } from '@routes/index';

export const App = (): JSX.Element => (
  <div>
    <Router>
      <h2>KnotSet.com</h2>
      <nav>
        <Link to="/">Home</Link> | <Link to="/login">Login</Link>
      </nav>

      <div>
        {routes.map((route, index) => (
          <Route key={index} {...route} />
        ))}
      </div>
    </Router>
  </div>
);

export default App;
