import React from 'react';
import styles from './home.module.css';
import bgImage from '../../assets/img/books.jpg';

export const Home: React.FC = () => {
  return (
    <div
      className={`contenair bg-cover w-full flex justify-center items-center ${styles.homeBackground}`}
      style={{ backgroundImage: `url(${bgImage})` }}
    >
      {/* Susan Q Yin */}
    </div>
  );
};

export default Home;
