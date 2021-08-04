import React from 'react';
import { BrowserRouter as Router, Link, Route } from 'react-router-dom';

import { routes } from '@routes/index';

export const App = (): JSX.Element => (
  <div className="text-gray-100">
    <Router>
      <h2>Hello App</h2>
      <nav>
        <b>Routing: </b>
        <Link to="/">Hello</Link>
        <Link to="/world">World</Link>
      </nav>
      <div>
        {routes.map((route, index) => (
          <Route key={index} {...route} />
        ))}
      </div>
    </Router>
  </div>
);
