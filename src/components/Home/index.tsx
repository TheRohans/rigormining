import React, { useState } from 'react';
import { Hello } from '@components/Home/Hello';

export const initialProps = {
  bar: 'First',
  foo: 'Page',
};

export default (): JSX.Element => {
  const [componentProps] = useState(initialProps);

  return <Hello {...componentProps} />;
};
