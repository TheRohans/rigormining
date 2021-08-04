import React from 'react';

type HelloProps = {
  foo: string;
  bar: string;
};
export const Hello: React.FC<HelloProps> = ({ foo, bar }) => {
  return (
    <div>
      Hello from {foo} and {bar}!
    </div>
  );
};
