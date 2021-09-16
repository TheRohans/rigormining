import React, { useState, useEffect } from 'react';
import Storage from '@aws-amplify/storage';

import { CloudDownloadIcon, EyeIcon } from '@heroicons/react/solid';

type ShelfListProps = {
  path: string;
  visitBook: (name: string, type: string) => void;
  downloadBookByName: (name: string) => void;
};

type LibraryItem = {
  name: string;
  type: string;
};

export const ShelfList: React.FC<ShelfListProps> = ({ path, visitBook, downloadBookByName }) => {
  const [booklist, setBookList] = useState<LibraryItem[]>([]);

  useEffect(() => {
    Storage.get(`${path}/metadata.json`, {
      level: 'public',
      download: true,
      contentType: 'application/json',
    })
      .then((v) => (v as any)?.Body?.text())
      .then((v) => {
        const json = JSON.parse(v);
        setBookList(json.children);
      })
      .catch((e) => {
        console.error(e);
      });
  }, []);

  const typeBadge = (type: string): JSX.Element => {
    switch (type) {
      case 'pdf':
      case 'png':
        return (
          <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">
            {type}
          </span>
        );
      case 'pub':
        return (
          <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-yellow-100 text-yellow-800">
            epub
          </span>
        );
    }
    return <span></span>;
  };

  return (
    <>
      {booklist.length > 0 && (
        <div className="-my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
          <div className="py-2 align-middle inline-block min-w-full sm:px-6 lg:px-8">
            <div className="shadow overflow-hidden border-b border-gray-200 sm:rounded-lg">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th
                      scope="col"
                      className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                    >
                      Name
                    </th>
                    <th
                      scope="col"
                      className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                    >
                      Type
                    </th>
                    <th scope="col" className="relative px-6 py-3">
                      <span className="sr-only">View</span>
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {booklist
                    .filter((i) => i.type !== undefined)
                    .map((book) => (
                      <tr key={book.name}>
                        <td
                          className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900"
                          onClick={() => visitBook(book.name, book.type)}
                        >
                          {book.name}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">{typeBadge(book.type)}</td>
                        <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                          <button
                            onClick={() => visitBook(book.name, book.type)}
                            className="text-indigo-600 mr-2 hover:text-indigo-900"
                          >
                            <EyeIcon className="h-4 w-4" aria-hidden="true" />
                          </button>

                          <button
                            onClick={() => downloadBookByName(book.name)}
                            className="text-indigo-600 hover:text-indigo-900"
                          >
                            <CloudDownloadIcon className="h-4 w-4" aria-hidden="true" />
                          </button>
                        </td>
                      </tr>
                    ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}
    </>
  );
};

export default ShelfList;
