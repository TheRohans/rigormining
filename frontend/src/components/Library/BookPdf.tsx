import React, { useRef, useState, useEffect } from 'react';
import { BookProps } from './Book';
import { ChevronLeftIcon, ChevronRightIcon, CloudDownloadIcon, XCircleIcon } from '@heroicons/react/solid';
import * as PDFJS from 'pdfjs-dist';
import { PDFDocumentProxy } from 'pdfjs-dist/types/src/display/api';

import styles from './bookpdf.module.css';
import Loader from './Loader';

const TOOLBAR_HEIGHT = 48;

export const BookPdf: React.FC<BookProps> = ({ type, url, title, loc, closeBook, downloadBook }) => {
  const refDisplay = useRef<HTMLCanvasElement>();
  const [currentPdf, setCurrentPdf] = useState<PDFDocumentProxy>();
  const [page, setPage] = useState<number>(1);
  const [numPages, setNumPages] = useState<number>(0);
  const [pageInput, setPageInput] = useState<string>('1');

  const [loading, setLoading] = useState<boolean>(true);

  PDFJS.GlobalWorkerOptions.workerSrc = 'pdf.worker.js';
  const task = PDFJS.getDocument(url);

  const goToPage = (target: number) => {
    const clamped = Math.min(Math.max(target, 1), numPages || target);
    setPage(clamped);
    setPageInput(String(clamped));
    if (currentPdf) renderPage(currentPdf, clamped);
  };

  const previousPage = () => goToPage(page - 1);
  const nextPage = () => goToPage(page + 1);

  const submitPageInput = () => {
    const target = parseInt(pageInput, 10);
    if (!isNaN(target)) goToPage(target);
    else setPageInput(String(page));
  };

  const keyListener = (e: KeyboardEvent) => {
    if (document.activeElement?.tagName === 'INPUT') return;
    if ((e.keyCode || e.which) == 37) previousPage();
    if ((e.keyCode || e.which) == 39) nextPage();
  };

  const renderPage = (pdf: PDFDocumentProxy, pageNum: number) => {
    ////////////
    const savedPages = localStorage.getItem('rm_pages');
    const spObj = JSON.parse(savedPages) ?? {};
    (spObj as any)[new URL(url).pathname] = pageNum;
    localStorage.setItem('rm_pages', JSON.stringify(spObj));
    ////////////

    pdf.getPage(pageNum).then((page) => {
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
        // Assign CSS to the textLayer element
        const textLayer = document.querySelector('.textLayer') as HTMLDivElement;
        textLayer.style.left = `${canvas.offsetLeft}px`;
        textLayer.style.top = `${canvas.offsetTop}px`;
        textLayer.style.height = `${canvas.height}px`;
        textLayer.style.width = `${canvas.width}px`;
        textLayer.style.position = 'absolute';

        textLayer.innerHTML = '';

        PDFJS.renderTextLayer({
          textContent: textContent,
          container: textLayer,
          viewport: viewport,
          textDivs: [],
        });
      });

      renderTask.promise.then(function () {
        // page rendered
      });
    });
  };

  useEffect(() => {
    ////////////
    const savedPages = localStorage.getItem('rm_pages');
    const spObj = JSON.parse(savedPages) ?? {};
    const pg = (spObj as any)[new URL(url).pathname] ?? 1;
    ////////////

    task.promise
      .then((pdf) => {
        setCurrentPdf(pdf);
        setNumPages(pdf.numPages);
        setPage(pg);
        setPageInput(String(pg));
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
              disabled={page <= 1}
              className="p-1.5 rounded-md text-gray-500 hover:text-gray-700 hover:bg-gray-100 disabled:opacity-30"
              title="Previous page"
            >
              <ChevronLeftIcon className="h-4 w-4" aria-hidden="true" />
            </button>
            <span className="text-sm text-gray-600 flex items-center gap-1">
              <input
                value={pageInput}
                onChange={(e) => setPageInput(e.target.value)}
                onBlur={submitPageInput}
                onKeyDown={(e) => e.key === 'Enter' && submitPageInput()}
                className="w-10 text-center border border-gray-300 rounded px-1 py-0.5 text-sm"
              />
              <span>of {numPages || '…'}</span>
            </span>
            <button
              type="button"
              onClick={nextPage}
              disabled={numPages > 0 && page >= numPages}
              className="p-1.5 rounded-md text-gray-500 hover:text-gray-700 hover:bg-gray-100 disabled:opacity-30"
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
        <div style={{ overflow: 'hidden' }} className="flex justify-center py-4 min-h-full relative">
          {loading && (
            <div className="flex items-center justify-center w-full">
              <Loader />
            </div>
          )}
          {!loading && (
            <>
              <canvas
                id="pdfarea"
                className={`rounded-sm shadow-md bg-white ${styles.bookDisplay}`}
                ref={refDisplay}
              ></canvas>
              <div className="textLayer"></div>
            </>
          )}
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

export default BookPdf;
