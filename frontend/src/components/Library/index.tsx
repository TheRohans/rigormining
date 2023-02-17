import React, { useState, useRef, useEffect, ChangeEvent } from 'react';
import { ShelfList } from './ShelfList';
import Book, { BookFormats } from './Book';
import Navigation from '@components/App/Navigation';
import { useUser } from '@components/App';
import { DocumentDisplayType } from '../Highlights/types';
import log from '@components/App/log';

export const Library: React.FC = () => {
  const [currentBook, setCurrentBook] = useState<string>('');
  const [currentBookType, setCurrentBookType] = useState<BookFormats>(undefined);

  const dropArea = useRef<HTMLDivElement>();
  const path = 'library/default';

  const viewBook = (name: string, contentType: string) => {
    log("View book: " + name);

    // TODO: hard coded URL
    setCurrentBook(
      `http://localhost:8000/public/library/rohan/${encodeURIComponent(name)}`
    );
  };

  const fetchBook = (name: string) => {
    log("Fetch book: " + name);
    // TODO: hard coded URL
    downloadBook(`http://localhost:8000/public/library/rohan/${encodeURIComponent(name)}`);
  };

  const visitBook = (name: string, type: DocumentDisplayType) => {
    log(`Type: ${type}`); 
    switch (type) {
      case 'pub':
        log("View epub: " + name);
        setCurrentBookType('epub');
        viewBook(name, 'application/epub+zip');
        break;
      case 'pdf':
        log("View pdf: " + name);
        setCurrentBookType('pdf');
        viewBook(name, 'application/pdf');
        break;
      default:
        alert("Yeah, nah. Try a pdf or epub");
    }
  };

  const closeBook = () => {
    setCurrentBook('');
    setCurrentBookType(undefined);
  };

  const downloadBook = (url: string) => {
    window.open(url, '_blank');
  };

  /////////////////////////////////////////////////////////////
  const uploadBook = async (name: string, file: File) => {
    console.log(`uploading ${name}`);

    try {
      const user = await useUser();
      const encoder = new TextEncoder();
      const data = encoder.encode(name);
      const hash = await crypto.subtle.digest('SHA-256', data);
      // Warning: this can't be await or it fails :-/
      // Storage.put(`${path}/${name}`, file, {
      //   level: 'public',
      // }).then((_v) => {
      //   console.log('done', _v);
      //   const d: Document = {
      //     id: `${user.username}:${btoa(String.fromCharCode.apply(null, new Uint8Array(hash)))}`,
      //     library: `${path}`,
      //     name: `${name}`,
      //     fileName: `${name}`,
      //     type: name.toLowerCase().endsWith('pdf') ? 'Pdf' : 'Epub',
      //   };
      //   API.graphql(graphqlOperation(createDocument, { input: d }));
      //   console.log('done done', d);
      // });
    } catch (e) {
      console.error(e);
    }
  };

  /////////////////////////////////////////////////////////////
  const uploadFiles = (e: ChangeEvent<HTMLInputElement>) => {
    Array.from(e.target?.files).forEach((f: File) => {
      console.log(f.name, f);
      uploadBook(f.name, f);
    });
  };

  const highlight = (e: DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const unhighlight = (e: DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleDrop = (e: DragEvent) => {
    e.preventDefault();
    e.stopPropagation();

    let dt = e.dataTransfer;
    let files = dt.files;

    Array.from(files).forEach((f) => {
      console.log(f.name, f);
      uploadBook(f.name, f);
    });
  };

  useEffect(() => {
    const listeners = new Map<string, EventListener>();

    if (dropArea?.current) {
      ['dragenter', 'dragover'].forEach((eventName) => {
        dropArea.current.addEventListener(eventName, highlight, false);
        listeners.set(eventName, highlight);
      });

      ['dragleave'].forEach((eventName) => {
        dropArea.current.addEventListener(eventName, unhighlight, false);
        listeners.set(eventName, unhighlight);
      });

      ['drop'].forEach((eventName) => {
        dropArea.current.addEventListener(eventName, handleDrop, false);
        listeners.set(eventName, handleDrop);
      });

      return () => {
        listeners.forEach((l, s) => {
          dropArea?.current?.removeEventListener(s, l);
        });
      };
    }
  });
  /////////////////////////////////////////////////////////////

  return (
    <>
      {currentBook.length <= 0 && (
        <>
          <Navigation loggedIn={true} />

          <div
            ref={dropArea}
            id="drop-area"
            className="border-2 border-gray-300 border-dashed p-2 m-3 max-w-7xl mx-auto sm:px-6 lg:px-8"
          >
            <form className="my-form">
              <p>
                Upload multiple documents with the file dialog, or by dragging and dropping documents onto the dashed
                region.
              </p>
              <input
                type="file"
                id="fileElem"
                multiple
                onChange={uploadFiles}
                accept=".pdf,.epub,application/pdf,application/epub,application/epub+zip,application/x-pdf"
              />
            </form>
          </div>

          <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
            <div className="py-5">
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
