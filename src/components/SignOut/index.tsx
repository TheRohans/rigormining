import React from 'react';
import { Auth } from 'aws-amplify';
import { useHistory } from 'react-router-dom';

export const SignOut: React.FC = () => {
  const history = useHistory();

  Auth.signOut().then(() => {
    localStorage.removeItem('isAuthenticated');
    history.push('/');
  });

  return <div>Signing out...</div>;
};

export default SignOut;
