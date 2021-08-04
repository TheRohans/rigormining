import { Icon } from '@components/Icon';
import React from 'react';

import Anchor from '../../assets/svg/anchor.svg';

type HomeProps = {
  foo: string;
  bar: string;
};

export const Home: React.FC<HomeProps> = ({ foo, bar }) => {
  return (
    <div className={'text-red-500'}>
      <Icon>
        <Anchor />
      </Icon>
      Hello from {foo ?? 'foo'} and {bar ?? 'bar'}!
    </div>
  );
};

export default Home;
