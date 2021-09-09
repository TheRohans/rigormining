// TODO: should get these from graphql?

export type Highlight = {
  id: string;
  type: 'Highlight' | 'Annotation';
  deviceType: 'Kindle' | 'Kobo';
  title: string;
  isbn?: string;
  author: string;
  page: number;
  startOffset?: number;
  text: string;
  annotation?: string;
  annotationExtra?: string;
  date: string;
};
export type HighlightList = Array<Highlight>;

export type DeviceList = {
  kindle?: number;
  kobo?: number;
};