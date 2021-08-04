import React, { useState } from 'react';
import { Auth } from 'aws-amplify';
import { useHistory } from 'react-router-dom';

// import FormElement from './FormElement';

export const SignIn = () => {
  const history = useHistory();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const signIn = (e: any) => {
    e.preventDefault();
    Auth.signIn({
      username: email,
      password,
    })
      .then((user) => {
        setEmail('');
        setPassword('');
        localStorage.setItem('isAuthenticated', JSON.stringify(true));
        // console.log(user);
        history.push('/dashboard');
      })
      .catch((err) => {
        console.log(err);
      });
  };

  return (
    <div className="form">
      <h3>Sign In</h3>
      <form>
        {/* <FormElement label="Email" forId="sign-in-email"> */}
        <input
          id="sign-in-email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="email"
        />
        {/* </FormElement> */}
        {/* <FormElement label="Password" forId="sign-in-password"> */}
        <input
          id="sign-in-password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="password"
        />
        {/* </FormElement> */}
        <button type="submit" onClick={signIn}>
          Sign In
        </button>
      </form>
    </div>
  );
};
export default SignIn;
