import React, { useState } from 'react';
import { SetS3Config } from '../../services';
import Storage from '@aws-amplify/storage';
import { ShelfList } from './ShelfList';
import Book, { BookFormats } from './Book';
import Navigation from '@components/App/Navigation';

export const Library: React.FC = () => {
  const [currentBook, setCurrentBook] = useState<string>('');
  const [currentBookType, setCurrentBookType] = useState<BookFormats>(undefined);

  SetS3Config('private.knotset.com', 'public');
  const path = 'library/rohan';

  const viewBook = (name: string, contentType: string) => {
    Storage.get(`${path}/${name}`, {
      level: 'public',
      download: false,
      contentType,
    })
      .then((v) => {
        console.log(v);
        setCurrentBook(v as string);
      })
      .catch((e) => {
        console.error(e);
      });
  };

  const fetchBook = (name: string) => {
    Storage.get(`${path}/${name}`, {
      level: 'public',
    })
      .then((v) => {
        downloadBook(v as string);
      })
      .catch((e) => {
        console.error(e);
      });
  };

  const visitBook = (name: string, type: string) => {
    switch (type) {
      case 'pub':
        setCurrentBookType('epub');
        viewBook(name, 'application/epub+zip');
        break;
      case 'pdf':
        setCurrentBookType('pdf');
        viewBook(name, 'application/pdf');
        // fetchBook(name);
        break;
    }
  };

  const closeBook = () => {
    setCurrentBook('');
    setCurrentBookType(undefined);
  };

  const downloadBook = (url: string) => {
    window.open(url, '_blank');
  };

  return (
    <>
      {currentBook.length <= 0 && (
        <>
          <Navigation loggedIn={true} />
          <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
            <div className="py-10">
              <main>
                <div className="flex flex-col">
                  {currentBook.length <= 0 && (
                    <ShelfList path={path} visitBook={visitBook} downloadBookByName={fetchBook} />
                  )}
                </div>
              </main>
            </div>
          </div>
        </>
      )}
      {currentBook.length > 0 && (
        <Book type={currentBookType} url={currentBook} closeBook={closeBook} downloadBook={downloadBook} />
      )}
    </>
  );
};

export default Library;
