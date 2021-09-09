import React, { useEffect, useState } from 'react';
import { API, graphqlOperation } from 'aws-amplify';
import { createHighlight, updateHighlight, deleteHighlight } from '../../graphql/mutations';
import { listHighlights } from '../../graphql/queries';
import { DeviceType, HighlightType } from '../../API';
import { Highlight, HighlightList, DeviceList } from './types';
import Navigation from '@components/App/Navigation';

declare const window: any;

export const Highlights: React.FC = () => {
  const [highlights, setHighlights] = useState<HighlightList>();

  const getHighlights = async () => {
    // const highlightDetails = {
    //   id: '12345',
    //   type: HighlightType.Highlight,
    //   deviceType: DeviceType.Pdf,
    //   title: 'My Crazy PDF',
    //   author: 'Rob Rohan',
    //   text: "This is some text that doesn't really exist",
    //   page: 3,
    // };

    // const newHighlight = await API.graphql({ query: createHighlight, variables: { input: highlightDetails } });

    const r = await API.graphql(graphqlOperation(listHighlights));
    setHighlights((r as any)?.data?.listHighlights?.items);
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
                <div className="ml-4 mt-4 flex-shrink-0 flex">
                  {/* <button
                  type="button"
                  className="relative inline-flex items-center px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  <PhoneIcon className="-ml-1 mr-2 h-5 w-5 text-gray-400" aria-hidden="true" />
                  <span>Phone</span>
                </button> */}
                </div>
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
                      <div className="w-full flex items-center justify-between p-6 space-x-6">
                        <div className="flex-1">
                          <div className="flex items-center space-x-3">
                            <h3 className="text-gray-900 text-sm font-medium truncate">{highlight.title}</h3>
                          </div>
                          <p className="mt-1 text-gray-500 text-sm whitespace-pre-line">{highlight.text}</p>
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
