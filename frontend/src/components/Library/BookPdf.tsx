import React, { useRef, useState, useEffect } from 'react';
import { BookProps } from './Book';
import { ChevronLeftIcon, ChevronRightIcon, CloudDownloadIcon, XCircleIcon } from '@heroicons/react/solid';
import * as PDFJS from 'pdfjs-dist';
import { PDFDocumentProxy } from 'pdfjs-dist/types/display/api';

import styles from './bookpdf.module.css';
import Loader from './Loader';

export const BookPdf: React.FC<BookProps> = ({ type, url, loc, closeBook, downloadBook }) => {
  const refDisplay = useRef<HTMLCanvasElement>();
  const [currentPdf, setCurrentPdf] = useState<PDFDocumentProxy>();
  const [page, setPage] = useState<number>(1);

  const [loading, setLoading] = useState<boolean>(true);

  PDFJS.GlobalWorkerOptions.workerSrc = 'pdf.worker.js';
  const task = PDFJS.getDocument(url);

  const previousPage = () => {
    try {
      setPage(page > 1 ? page - 1 : 1);
      renderPage(currentPdf, page);
    } catch (e) {
      // Page probably too far
      console.warn('Previous page: ', e);
    }
  };

  const nextPage = () => {
    try {
      setPage(page + 1);
      renderPage(currentPdf, page);
    } catch (e) {
      // Page probably too far
      console.warn('Next page: ', e);
    }
  };

  const keyListener = (e: KeyboardEvent) => {
    const MOD = e.ctrlKey || e.metaKey;
    if ((e.keyCode || e.which) == 37) previousPage();
    if ((e.keyCode || e.which) == 39) nextPage();
  };

  const renderPage = (pdf: PDFDocumentProxy, pageNum: number) => {
    ////////////
    const savedPages = localStorage.getItem('ks_pages');
    const spObj = JSON.parse(savedPages) ?? {};
    (spObj as any)[new URL(url).pathname] = pageNum;
    localStorage.setItem('ks_pages', JSON.stringify(spObj));
    ////////////

    pdf.getPage(pageNum).then((page) => {
      // console.log('Page loaded');

      const scale = 1.1;
      const viewport = page.getViewport({ scale: scale });

      // Prepare canvas using PDF page dimensions
      const canvas = document.getElementById('pdfarea') as HTMLCanvasElement;
      const context = canvas.getContext('2d');
      canvas.height = viewport.height;
      canvas.width = viewport.width;

      // Render PDF page into canvas context
      const renderContext = {
        canvasContext: context,
        viewport: viewport,
      };
      const renderTask = page.render(renderContext);

      page.getTextContent().then((textContent) => {
        // console.log(textContent);

        // Assign CSS to the textLayer element
        const textLayer = document.querySelector('.textLayer') as HTMLDivElement;
        textLayer.style.left = `${canvas.offsetLeft}px`;
        textLayer.style.top = `${canvas.offsetTop}px`;
        // textLayer.style.height = canvas.offsetHeight + 'px';
        // textLayer.style.width = canvas.offsetWidth + 'px';
        textLayer.style.height = `${canvas.height}px`;
        textLayer.style.width = `${canvas.width}px`;
        textLayer.style.position = 'absolute';

        textLayer.innerHTML = '';

        // Pass the data to the method for rendering of text over the pdf canvas.
        PDFJS.renderTextLayer({
          textContent: textContent,
          container: textLayer,
          viewport: viewport,
          textDivs: [],
        });

        //  const textLayer = new TextLayerBuilder({
        //      textLayerDiv : $textLayerDiv.get(0),
        //      pageIndex : page_num - 1,
        //      viewport : viewport
        //  });

        //  textLayer.setTextContent(textContent);
        //  textLayer.render();
      });

      renderTask.promise.then(function () {
        // console.log('Page rendered');
      });
    });
  };

  useEffect(() => {
    ////////////
    const savedPages = localStorage.getItem('ks_pages');
    const spObj = JSON.parse(savedPages) ?? {};
    const pg = (spObj as any)[new URL(url).pathname] ?? 1;
    ////////////

    task.promise
      .then((pdf) => {
        // console.log('pdf loaded');
        setCurrentPdf(pdf);
        setPage(pg);
        renderPage(pdf, pg);
        setLoading(false);
      })
      .catch((err) => {
        console.error(err);
      });

    document.addEventListener('keyup', keyListener, false);

    return () => {
      document.removeEventListener('keyup', keyListener);
    };
  }, []);

  return (
    <>
      <main>
        <div style={{ overflow: 'hidden' }} className="flex justify-center py-2 sm:px-0 max-h-full relative">
          {loading && <Loader />}
          {!loading && (
            <>
              <canvas id="pdfarea" className={`rounded-lg ${styles.bookDisplay} max-h-full`} ref={refDisplay}></canvas>
              <div className="textLayer"></div>
            </>
          )}
        </div>

        <div
          className="leftSide h-full w-2/12 fixed left-0 top-0 bg-gray-300 bg-opacity-0 cursor-pointer"
          onClick={() => previousPage()}
        ></div>
        <div
          className="rightSide h-full w-2/12 fixed right-0 top-0 bg-gray-300 bg-opacity-0 cursor-pointer"
          onClick={() => nextPage()}
        ></div>

        <div className="fixed top-0 left-0 w-full">
          {/*  max-w-7xl mx-auto */}
          <div className="flex justify-end">
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

            {/* <span className="relative z-0 inline-flex shadow-sm rounded-md">
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
            </span> */}
          </div>
        </div>
      </main>
    </>
  );
};

export default BookPdf;
