import React, { useEffect, useState, useRef } from 'react';
import { BookProps } from './Book';
import ePub, { Book, Rendition } from 'epubjs';
import { ChevronLeftIcon, ChevronRightIcon, CloudDownloadIcon, XCircleIcon } from '@heroicons/react/solid';

import styles from './bookepub.module.css';
import Loader from './Loader';

const TOOLBAR_HEIGHT = 48;

export const BookEpub: React.FC<BookProps> = ({ type, url, title, loc, closeBook, downloadBook }) => {
  const [rendition, setRendition] = useState<Rendition>();
  const [loading, setLoading] = useState<boolean>(true);

  const refDisplay = useRef<HTMLDivElement>();
  const currentSectionIndex = loc ? loc : undefined;

  const previousPage = () => {
    rendition?.prev();
  };

  const nextPage = () => {
    rendition?.next();
  };

  const keyListener = (e: KeyboardEvent) => {
    if ((e.keyCode || e.which) == 37) previousPage();
    if ((e.keyCode || e.which) == 39) nextPage();
    if ((e.keyCode || e.which) == 88) closeBook();
  };

  const selectedRange = function (book: Book) {
    return function (cfiRange: any) {
      // The API typescript seems to have the wrong types
      (book.getRange(cfiRange) as unknown as Promise<Range>).then((r) => {
        console.log(r.toString());
      });
    };
  };

  useEffect(() => {
    refDisplay.current.textContent = '';
    const book = ePub(url, {
      openAs: type,
    });

    const rendLocal = book.renderTo(refDisplay.current.id, {
      manager: 'continuous',
      flow: 'paginated',
      width: '100%',
      height: '100%',
    });

    const displayed = rendLocal.display(currentSectionIndex);

    displayed.then(() => {
      setRendition(rendLocal);
    });

    book.ready.then(() => {
      rendLocal.on('keyup', keyListener);
      rendLocal.on('selected', selectedRange(book));

      setLoading(false);
    });

    document.addEventListener('keyup', keyListener, false);

    return () => {
      document.removeEventListener('keyup', keyListener);
    };
  }, [url]);

  return (
    <div className="fixed inset-0 bg-gray-200 flex flex-col">
      <div
        className="flex items-center justify-between gap-3 px-3 bg-white border-b border-gray-200 shadow-sm flex-shrink-0 z-50"
        style={{ height: TOOLBAR_HEIGHT }}
      >
        <div className="flex items-center gap-2 min-w-0">
          <button
            type="button"
            onClick={closeBook}
            className="p-1.5 rounded-md text-gray-500 hover:text-gray-700 hover:bg-gray-100"
            title="Back to library"
          >
            <XCircleIcon className="h-5 w-5" aria-hidden="true" />
          </button>
          {title && <span className="text-sm font-medium text-gray-700 truncate">{title}</span>}
        </div>

        {!loading && (
          <div className="flex items-center gap-1 flex-shrink-0">
            <button
              type="button"
              onClick={previousPage}
              className="p-1.5 rounded-md text-gray-500 hover:text-gray-700 hover:bg-gray-100"
              title="Previous page"
            >
              <ChevronLeftIcon className="h-4 w-4" aria-hidden="true" />
            </button>
            <button
              type="button"
              onClick={nextPage}
              className="p-1.5 rounded-md text-gray-500 hover:text-gray-700 hover:bg-gray-100"
              title="Next page"
            >
              <ChevronRightIcon className="h-4 w-4" aria-hidden="true" />
            </button>
          </div>
        )}

        <button
          type="button"
          onClick={() => downloadBook(url)}
          className="p-1.5 rounded-md text-gray-500 hover:text-gray-700 hover:bg-gray-100 flex-shrink-0"
          title="Download"
        >
          <CloudDownloadIcon className="h-5 w-5" aria-hidden="true" />
        </button>
      </div>

      <main className="flex-1 relative overflow-auto">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-4 h-full">
          {loading && (
            <div className="flex items-center justify-center w-full h-full">
              <Loader />
            </div>
          )}
          <div
            id="bookarea"
            className={`rounded-sm bg-white shadow-md h-full ${styles.bookDisplay}`}
            ref={refDisplay}
          ></div>
        </div>

        {!loading && (
          <>
            <div
              className="fixed left-0 w-1/6 bg-transparent cursor-pointer"
              style={{ top: TOOLBAR_HEIGHT, bottom: 0 }}
              onClick={previousPage}
            ></div>
            <div
              className="fixed right-0 w-1/6 bg-transparent cursor-pointer"
              style={{ top: TOOLBAR_HEIGHT, bottom: 0 }}
              onClick={nextPage}
            ></div>
          </>
        )}
      </main>
    </div>
  );
};

export default BookEpub;
