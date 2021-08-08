import React from 'react';
import BookEpub from './BookEpub';

type BookProps = {
  type: 'epub' | 'pdf';
  url: string;
  loc?: number;
  closeBook: () => void;
  downloadBook: (url: string) => void;
};

export const Book: React.FC<BookProps> = ({ type, url, loc, closeBook, downloadBook }) => {
  return <BookEpub type="epub" url={url} closeBook={closeBook} downloadBook={downloadBook} />;
};

export default Book;
