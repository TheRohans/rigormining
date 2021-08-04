import React from 'react';
import classnames from 'classnames';
import styles from './icon.module.css';

type IconProps = {
  style?: 'small' | 'standard' | 'large';
};

export const Icon: React.FC<IconProps> = (props) => {
  return (
    <div
      className={classnames({
        [styles.icon]: true,
        [styles.large]: props.style === 'large',
        [styles.standard]: props.style === 'standard' || props.style === undefined,
        [styles.small]: props.style === 'small',
      })}
    >
      {props.children}
    </div>
  );
};
