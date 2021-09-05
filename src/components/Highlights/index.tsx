import React, { useEffect, useState } from 'react';
import { API, graphqlOperation } from 'aws-amplify';
import { createHighlight, updateHighlight, deleteHighlight } from '../../graphql/mutations';
import { listHighlights } from '../../graphql/queries';
import { DeviceType, HighlightType } from '../../API';
// import { Highlight } from '../../API';

export const Highlights: React.FC = () => {
  const [highlights, setHighlights] = useState([]);

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
    // console.log(r);
    setHighlights((r as any)?.data?.listHighlights?.items);
  };

  useEffect(() => {
    getHighlights();
  }, []);

  return (
    <div>
      {highlights.length &&
        highlights.map((highlight, i) => {
          return <div key={highlight.id}>{highlight.title}</div>;
        })}
    </div>
  );
};

export default Highlights;
