import React from 'react';
import { loginUrl, devLoginUrl } from '../../api/client';

// Only set in frontend/Makefile's make_dev_env, so this stays out of the
// production bundle (make_prod_env doesn't set it) even though the server
// already 404s /dev-login unless RM_AUTH_DEV_LOGIN=true is set too.
const devLoginEnabled = process.env.RIGORMINING_DEV_LOGIN === 'true';

export const SignIn: React.FC = () => {
  return (
    <div className="flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8 min-h-screen">
      <div className="max-w-md w-full space-y-8 text-center">
        <h2 className="text-3xl font-extrabold text-gray-900">Sign in to your library</h2>
        <a
          href={loginUrl}
          className="group relative w-full inline-flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
        >
          Sign in with Google
        </a>
        {devLoginEnabled && (
          <a href={devLoginUrl} className="block text-xs text-gray-400 hover:text-gray-600">
            Dev login (local only)
          </a>
        )}
      </div>
    </div>
  );
};

export default SignIn;
