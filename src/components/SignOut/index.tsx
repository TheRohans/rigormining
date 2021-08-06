import React from 'react';
import { useHistory } from 'react-router-dom';

type SignOutProps = {
  logOut: () => Promise<void>;
};

export const SignOut: React.FC<SignOutProps> = ({ logOut }) => {
  const history = useHistory();

  logOut().then(() => {
    history.push('/');
  });

  return <div>Signing out...</div>;
};

export default SignOut;
