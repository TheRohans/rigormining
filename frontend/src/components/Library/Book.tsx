import React from 'react';
import BookEpub from './BookEpub';
import BookPdf from './BookPdf';

export type BookFormats = 'epub' | 'pdf';
export type BookProps = {
  type: BookFormats;
  url: string;
  title?: string;
  loc?: number;
  closeBook: () => void;
  downloadBook: (url: string) => void;
};

export const Book: React.FC<BookProps> = ({ type, url, title, loc, closeBook, downloadBook }) => {
  return (
    <>
      {(type as string) === 'pdf' && (
        <BookPdf type={type} url={url} title={title} closeBook={closeBook} downloadBook={downloadBook} />
      )}
      {(type as string) === 'epub' && (
        <BookEpub type={type} url={url} title={title} closeBook={closeBook} downloadBook={downloadBook} />
      )}
    </>
  );
};

export default Book;
