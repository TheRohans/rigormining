import React, { useEffect, useState } from 'react';
// import { API, graphqlOperation } from 'aws-amplify';
// import { deleteHighlight } from '../../graphql/mutations';
// import { listHighlights } from '../../graphql/queries';
import { HighlightList } from './types';
import Navigation from '@components/App/Navigation';

declare const window: any;

export const Highlights: React.FC = () => {
  const [highlights, setHighlights] = useState<HighlightList>();

  const getHighlights = async () => {
    // const r = await API.graphql(graphqlOperation(listHighlights));
    const r = {};
    setHighlights((r as any)?.data?.listHighlights?.items);
  };

  const removeHighlight = async (id: string) => {
    try {
      // await API.graphql(graphqlOperation(deleteHighlight, { input: { id } }));
      setHighlights(highlights.filter((h) => h.id !== id));
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    getHighlights();
  }, []);

  return (
    <>
      <Navigation loggedIn={true} />
      <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
        <div className="py-10">
          <main>
            <div className="bg-white px-4 py-5 border border-gray-200 sm:px-6 mb-2 ">
              <div className="-ml-4 -mt-4 flex justify-between items-center flex-wrap sm:flex-nowrap">
                <div className="ml-4 mt-4">
                  <div className="flex items-center">
                    <div className="ml-4">
                      <h3 className="text-lg leading-6 font-medium text-gray-900">Highlights</h3>
                    </div>
                  </div>
                </div>
                <div className="ml-4 mt-4 flex-shrink-0 flex"></div>
              </div>
            </div>

            <ul role="list" className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {highlights &&
                highlights.map((highlight, i) => {
                  return (
                    <li
                      key={`${highlight.id}${highlight.type}`}
                      className="col-span-1 bg-white rounded-lg shadow divide-y divide-gray-200"
                    >
                      <div className="w-full h-full flex items-center justify-between p-6 space-x-6">
                        <div className="w-full h-full flex justify-between flex-col">
                          <div className="flex items-center space-x-3 truncate">
                            <h3 className="text-gray-900 text-sm font-medium truncate overflow-ellipsis">
                              {highlight.title}
                            </h3>
                          </div>

                          <p className="mt-1 text-gray-500 text-sm flex-grow">{highlight.text}</p>

                          <div className="w-full flex divide-x divide-gray-200 justify-evenly mt-4">
                            <span className="relative z-0 inline-flex shadow-sm rounded-md">
                              <button
                                type="button"
                                onClick={() => removeHighlight(highlight.id)}
                                className="-ml-px relative inline-flex items-center px-4 py-2 border border-gray-300 bg-red-300 text-sm font-medium text-gray-300 hover:bg-red-500 focus:z-10 focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500"
                              >
                                Delete
                              </button>

                              {/* <button
                                type="button"
                                onClick={() => removeHighlight(highlight.id)}
                                className="relative inline-flex items-center px-4 py-2 rounded-l-md border border-gray-300 bg-red-300 text-sm font-medium text-white  hover:bg-red-500 focus:z-10 focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500"
                              >
                                Delete
                              </button> */}
                              {/* <button
                                type="button"
                                className="-ml-px relative inline-flex items-center px-4 py-2 border border-gray-300 bg-white text-sm font-medium text-gray-300 hover:bg-gray-50 focus:z-10 focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500"
                              >
                                Reference
                              </button>
                              <button
                                type="button"
                                className="-ml-px relative inline-flex items-center px-4 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-300 hover:bg-gray-50 focus:z-10 focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500"
                              >
                                Share
                              </button> */}
                            </span>
                          </div>
                        </div>
                      </div>
                    </li>
                  );
                })}
            </ul>
          </main>
        </div>
      </div>
    </>
  );
};

export default Highlights;
