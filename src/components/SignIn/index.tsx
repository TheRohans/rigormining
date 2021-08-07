import React, { useState } from 'react';
import { useHistory } from 'react-router-dom';
import { LockClosedIcon, RefreshIcon } from '@heroicons/react/solid';

import Error from '../Alerts/Error';

type SignInProps = {
  logIn: (email: string, password: string) => Promise<any>;
};

export const SignIn: React.FC<SignInProps> = ({ logIn }) => {
  const [working, setWorking] = useState(false);
  const [error, setError] = useState('');

  const history = useHistory();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const signIn = (e: any): boolean => {
    e.preventDefault();
    setWorking(true);
    setError('');
    logIn(email, password)
      .then((u) => {
        setEmail('');
        setPassword('');
        history.push('/dashboard');
        setWorking(false);
      })
      .catch((e) => {
        console.error(e);
        setError(e.message);
        setWorking(false);
      });

    return false;
  };

  return (
    <div className="flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <img
            className="mx-auto h-12 w-auto"
            src="https://tailwindui.com/img/logos/workflow-mark-indigo-600.svg"
            alt="Workflow"
          />
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">Sign in to your account</h2>
        </div>

        {error.length > 0 && <Error text={error} />}

        <form className="mt-8 space-y-6" onSubmit={signIn}>
          <input type="hidden" name="remember" defaultValue="true" />
          <div className="rounded-md shadow-sm -space-y-px">
            <div>
              <label htmlFor="email-address" className="sr-only">
                Email address
              </label>
              <input
                id="email-address"
                name="email"
                type="email"
                autoComplete="email"
                required
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                placeholder="Email address"
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
            <div>
              <label htmlFor="password" className="sr-only">
                Password
              </label>
              <input
                id="password"
                name="password"
                type="password"
                autoComplete="current-password"
                required
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                placeholder="Password"
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
          </div>

          <div className="flex items-center justify-between">
            <div className="flex items-center">
              {/* <input
                id="remember-me"
                name="remember-me"
                type="checkbox"
                className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
              />
              <label htmlFor="remember-me" className="ml-2 block text-sm text-gray-900">
                Remember me
              </label> */}
            </div>

            <div className="text-sm text-gray-900">
              {/* <a href="#" className="font-medium text-indigo-600 hover:text-indigo-500">
                Forgot your password?
              </a> */}
              version: {process.env.KNOTSET_VERSION}
            </div>
          </div>

          <div>
            <button
              type="submit"
              className={`group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 ${
                working ? 'cursor-wait' : ''
              }`}
              disabled={working}
            >
              <span className="absolute left-0 inset-y-0 flex items-center pl-3">
                {!working && (
                  <LockClosedIcon className="h-5 w-5 text-indigo-500 group-hover:text-indigo-400" aria-hidden="true" />
                )}
                {working && (
                  <RefreshIcon
                    className="h-5 w-5 text-indigo-500 group-hover:text-indigo-400 animate-spin"
                    aria-hidden="true"
                  />
                )}
              </span>
              Sign in
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default SignIn;
