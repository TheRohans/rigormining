import * as React from 'react';
import * as ReactDOM from 'react-dom';

import App from './components/App/index';

// to get tailwind
import './assets/css/style.css';

import { configureAmplify } from './services';

configureAmplify();
ReactDOM.render(<App />, document.getElementById('root'));
