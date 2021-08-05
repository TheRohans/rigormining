import React, { useState } from 'react';
import { Auth } from 'aws-amplify';
import { useHistory } from 'react-router-dom';

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
    <form>
      <div className="bg-white shadow-md rounded px-8 pt-6 pb-8 mb-4 flex flex-col">
        <div className="mb-4">
          <label className="block text-grey-darker text-sm font-bold mb-2">Username</label>
          <input
            className="shadow appearance-none border rounded w-full py-2 px-3 text-grey-darker"
            id="sign-in-email"
            type="email"
            value={email}
            placeholder="email"
            onChange={(e) => setEmail(e.target.value)}
          />
        </div>
        <div className="mb-6">
          <label className="block text-grey-darker text-sm font-bold mb-2">Password</label>
          <input
            className="shadow appearance-none border border-red rounded w-full py-2 px-3 text-grey-darker mb-3"
            id="sign-in-password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="password"
          />
          <p className="text-red text-xs italic">Please choose a password.</p>
        </div>
        <div className="flex items-center justify-between">
          <button
            className="bg-blue-800 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
            type="button"
            onClick={signIn}
          >
            Sign In
          </button>
          {/* <a className="inline-block align-baseline font-bold text-sm text-blue hover:text-blue-darker" href="#">
            Forgot Password?
          </a> */}
        </div>
      </div>
    </form>
  );
};
export default SignIn;

// {/* <div className="form">
//   <h3>Sign In</h3>
//   <form>
//     <input
//       id="sign-in-email"
//       type="email"
//       value={email}
//       onChange={(e) => setEmail(e.target.value)}
//       placeholder="email"
//     />
//     <input
//       id="sign-in-password"
//       type="password"
//       value={password}
//       onChange={(e) => setPassword(e.target.value)}
//       placeholder="password"
//     />
//     <button type="submit" onClick={signIn}>
//       Sign In
//     </button>
//   </form>
// </div> */}
