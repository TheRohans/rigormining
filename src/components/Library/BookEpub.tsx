import React, { useEffect, useState, useRef } from 'react';
import { BookProps } from './Book';
import ePub, { Rendition } from 'epubjs';
import { ChevronLeftIcon, ChevronRightIcon, CloudDownloadIcon, XCircleIcon } from '@heroicons/react/solid';

import styles from './bookepub.module.css';
import Loader from './Loader';

export const BookEpub: React.FC<BookProps> = ({ type, url, loc, closeBook, downloadBook }) => {
  const [rendition, setRendition] = useState<Rendition>();
  const [loading, setLoading] = useState<boolean>(true);

  const refDisplay = useRef<HTMLDivElement>();
  var currentSectionIndex = loc ? loc : undefined;

  const previousPage = () => {
    rendition?.prev();
  };

  const nextPage = () => {
    rendition?.next();
  };

  const keyListener = (e: KeyboardEvent) => {
    const MOD = e.ctrlKey || e.metaKey;
    if ((e.keyCode || e.which) == 37) previousPage();
    if ((e.keyCode || e.which) == 39) nextPage();
  };

  // const addBookmark = function (cfi: any) {
  //   var present = this.isBookmarked(cfi);
  //   if (present > -1) return;

  //   this.settings.bookmarks.push(cfi);

  //   this.trigger('reader:bookmarked', cfi);
  // };

  // const removeBookmark = function (cfi: any) {
  //   var bookmark = this.isBookmarked(cfi);
  //   if (bookmark === -1) return;

  //   this.settings.bookmarks.splice(bookmark, 1);

  //   this.trigger('reader:unbookmarked', bookmark);
  // };

  const selectedRange = function (cfiRange: any) {
    const cfiFragment = '#' + cfiRange;
    console.log(cfiFragment);
    // // Update the History Location
    // if(this.settings.history && window.location.hash != cfiFragment) {
    //   // Add CFI fragment to the history
    //   history.pushState({}, '', cfiFragment);
    //   this.currentLocationCfi = cfiRange;
    // }
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

    displayed.then((renderer) => {
      console.log('display done', renderer);
      setRendition(rendLocal);
    });

    // Navigation loaded
    book.loaded.navigation.then((toc) => {
      console.log('toc', toc);
    });

    book.ready.then(() => {
      rendLocal.on('keyup', keyListener);
      rendLocal.on('selected', selectedRange);

      setLoading(false);
    });

    document.addEventListener('keyup', keyListener, false);

    return () => {
      document.removeEventListener('keyup', keyListener);
    };
  }, [url]);

  // kind of dodgy

  return (
    <div>
      <header>
        <div className="flex justify-center max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <h1 className="text-3xl font-bold leading-tight text-gray-900">
            <button
              type="button"
              className="inline-flex items-center mx-2 px-2 py-2 border border-transparent shadow-sm text-sm leading-4 font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              onClick={closeBook}
            >
              <XCircleIcon className="h-4 w-4" aria-hidden="true" />
            </button>
            {/* <button
              type="button"
              className="inline-flex items-center mx-2 px-2 py-2 border border-transparent shadow-sm text-sm leading-4 font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              onClick={() => downloadBook(url)}
            >
              <CloudDownloadIcon className="h-4 w-4" aria-hidden="true" />
            </button> */}

            <span className="relative z-0 inline-flex shadow-sm rounded-md">
              <button
                type="button"
                className="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 focus:z-10 focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500"
                onClick={() => previousPage()}
              >
                <span className="sr-only">Previous</span>
                <ChevronLeftIcon className="h-4 w-4" aria-hidden="true" />
              </button>
              <button
                type="button"
                className="-ml-px relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 focus:z-10 focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500"
                onClick={() => nextPage()}
              >
                <span className="sr-only">Next</span>
                <ChevronRightIcon className="h-4 w-4" aria-hidden="true" />
              </button>
            </span>
          </h1>
        </div>
      </header>
      <main>
        <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
          <div className="px-2 py-2 sm:px-0">
            {loading && <Loader />}
            <div id="bookarea" className={`rounded-lg ${styles.bookDisplay}`} ref={refDisplay}></div>
          </div>
        </div>
      </main>
    </div>
  );
};

export default BookEpub;
