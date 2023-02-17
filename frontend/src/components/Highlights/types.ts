// TODO: should get these from graphql?

export type DeviceType = 'Kindle' | 'Kobo' | 'Pdf' | 'Epub' | 'Web';

export type Highlight = {
  id: string;
  type: 'Highlight' | 'Annotation';
  deviceType: DeviceType;
  title: string;
  isbn?: string;
  author: string;
  page: number;
  startOffset?: number;
  text: string;
  annotation?: string;
  annotationExtra?: string;
  date?: string;
  uri?: string;
};
export type HighlightList = Array<Highlight>;

export type DeviceList = {
  kindle?: number;
  kobo?: number;
};

export type DocumentDisplayType = 'pdf' | 'pub' | 'png' | 'jpg'

export type Document = {
  id: string;
  library: string;
  name: string;
  fileName: string;
  type: DocumentDisplayType;
};
