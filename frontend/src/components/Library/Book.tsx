import React from 'react';
import BookEpub from './BookEpub';
import BookPdf from './BookPdf';

export type BookFormats = 'epub' | 'pdf';
export type BookProps = {
  type: BookFormats;
  url: string;
  loc?: number;
  closeBook: () => void;
  downloadBook: (url: string) => void;
};

export const Book: React.FC<BookProps> = ({ type, url, loc, closeBook, downloadBook }) => {
  return (
    <>
      {(type as String) === 'pdf' && (
        <BookPdf type={type} url={url} closeBook={closeBook} downloadBook={downloadBook} />
      )}
      {(type as String) === 'epub' && (
        <BookEpub type={type} url={url} closeBook={closeBook} downloadBook={downloadBook} />
      )}
    </>
  );
};

export default Book;
