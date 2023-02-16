/* tslint:disable */
/* eslint-disable */
//  This file was automatically generated and should not be edited.

export type CreateHighlightInput = {
  id?: string | null,
  type: HighlightType,
  deviceType: DeviceType,
  title: string,
  isbn?: string | null,
  author?: string | null,
  page?: number | null,
  startOffset?: number | null,
  endOffset?: number | null,
  text?: string | null,
  annotation?: string | null,
  annotationExtra?: string | null,
  date?: string | null,
  uri?: string | null,
};

export enum HighlightType {
  Annotation = "Annotation",
  Highlight = "Highlight",
}


export enum DeviceType {
  Kindle = "Kindle",
  Kobo = "Kobo",
  Pdf = "Pdf",
  Epub = "Epub",
}


export type ModelHighlightConditionInput = {
  type?: ModelHighlightTypeInput | null,
  deviceType?: ModelDeviceTypeInput | null,
  title?: ModelStringInput | null,
  isbn?: ModelStringInput | null,
  author?: ModelStringInput | null,
  page?: ModelIntInput | null,
  startOffset?: ModelIntInput | null,
  endOffset?: ModelIntInput | null,
  text?: ModelStringInput | null,
  annotation?: ModelStringInput | null,
  annotationExtra?: ModelStringInput | null,
  date?: ModelStringInput | null,
  uri?: ModelStringInput | null,
  and?: Array< ModelHighlightConditionInput | null > | null,
  or?: Array< ModelHighlightConditionInput | null > | null,
  not?: ModelHighlightConditionInput | null,
};

export type ModelHighlightTypeInput = {
  eq?: HighlightType | null,
  ne?: HighlightType | null,
};

export type ModelDeviceTypeInput = {
  eq?: DeviceType | null,
  ne?: DeviceType | null,
};

export type ModelStringInput = {
  ne?: string | null,
  eq?: string | null,
  le?: string | null,
  lt?: string | null,
  ge?: string | null,
  gt?: string | null,
  contains?: string | null,
  notContains?: string | null,
  between?: Array< string | null > | null,
  beginsWith?: string | null,
  attributeExists?: boolean | null,
  attributeType?: ModelAttributeTypes | null,
  size?: ModelSizeInput | null,
};

export enum ModelAttributeTypes {
  binary = "binary",
  binarySet = "binarySet",
  bool = "bool",
  list = "list",
  map = "map",
  number = "number",
  numberSet = "numberSet",
  string = "string",
  stringSet = "stringSet",
  _null = "_null",
}


export type ModelSizeInput = {
  ne?: number | null,
  eq?: number | null,
  le?: number | null,
  lt?: number | null,
  ge?: number | null,
  gt?: number | null,
  between?: Array< number | null > | null,
};

export type ModelIntInput = {
  ne?: number | null,
  eq?: number | null,
  le?: number | null,
  lt?: number | null,
  ge?: number | null,
  gt?: number | null,
  between?: Array< number | null > | null,
  attributeExists?: boolean | null,
  attributeType?: ModelAttributeTypes | null,
};

export type UpdateHighlightInput = {
  id: string,
  type?: HighlightType | null,
  deviceType?: DeviceType | null,
  title?: string | null,
  isbn?: string | null,
  author?: string | null,
  page?: number | null,
  startOffset?: number | null,
  endOffset?: number | null,
  text?: string | null,
  annotation?: string | null,
  annotationExtra?: string | null,
  date?: string | null,
  uri?: string | null,
};

export type DeleteHighlightInput = {
  id?: string | null,
};

export type CreateDocumentInput = {
  id?: string | null,
  library: string,
  name: string,
  fileName: string,
  type?: DeviceType | null,
};

export type ModelDocumentConditionInput = {
  library?: ModelStringInput | null,
  name?: ModelStringInput | null,
  fileName?: ModelStringInput | null,
  type?: ModelDeviceTypeInput | null,
  and?: Array< ModelDocumentConditionInput | null > | null,
  or?: Array< ModelDocumentConditionInput | null > | null,
  not?: ModelDocumentConditionInput | null,
};

export type UpdateDocumentInput = {
  id: string,
  library?: string | null,
  name?: string | null,
  fileName?: string | null,
  type?: DeviceType | null,
};

export type DeleteDocumentInput = {
  id?: string | null,
};

export type ModelHighlightFilterInput = {
  id?: ModelIDInput | null,
  type?: ModelHighlightTypeInput | null,
  deviceType?: ModelDeviceTypeInput | null,
  title?: ModelStringInput | null,
  isbn?: ModelStringInput | null,
  author?: ModelStringInput | null,
  page?: ModelIntInput | null,
  startOffset?: ModelIntInput | null,
  endOffset?: ModelIntInput | null,
  text?: ModelStringInput | null,
  annotation?: ModelStringInput | null,
  annotationExtra?: ModelStringInput | null,
  date?: ModelStringInput | null,
  uri?: ModelStringInput | null,
  and?: Array< ModelHighlightFilterInput | null > | null,
  or?: Array< ModelHighlightFilterInput | null > | null,
  not?: ModelHighlightFilterInput | null,
};

export type ModelIDInput = {
  ne?: string | null,
  eq?: string | null,
  le?: string | null,
  lt?: string | null,
  ge?: string | null,
  gt?: string | null,
  contains?: string | null,
  notContains?: string | null,
  between?: Array< string | null > | null,
  beginsWith?: string | null,
  attributeExists?: boolean | null,
  attributeType?: ModelAttributeTypes | null,
  size?: ModelSizeInput | null,
};

export type ModelDocumentFilterInput = {
  id?: ModelIDInput | null,
  library?: ModelStringInput | null,
  name?: ModelStringInput | null,
  fileName?: ModelStringInput | null,
  type?: ModelDeviceTypeInput | null,
  and?: Array< ModelDocumentFilterInput | null > | null,
  or?: Array< ModelDocumentFilterInput | null > | null,
  not?: ModelDocumentFilterInput | null,
};

export type CreateHighlightMutationVariables = {
  input: CreateHighlightInput,
  condition?: ModelHighlightConditionInput | null,
};

export type CreateHighlightMutation = {
  createHighlight:  {
    __typename: "Highlight",
    id: string,
    type: HighlightType,
    deviceType: DeviceType,
    title: string,
    isbn: string | null,
    author: string | null,
    page: number | null,
    startOffset: number | null,
    endOffset: number | null,
    text: string | null,
    annotation: string | null,
    annotationExtra: string | null,
    date: string | null,
    uri: string | null,
    createdAt: string,
    updatedAt: string,
    owner: string | null,
  } | null,
};

export type UpdateHighlightMutationVariables = {
  input: UpdateHighlightInput,
  condition?: ModelHighlightConditionInput | null,
};

export type UpdateHighlightMutation = {
  updateHighlight:  {
    __typename: "Highlight",
    id: string,
    type: HighlightType,
    deviceType: DeviceType,
    title: string,
    isbn: string | null,
    author: string | null,
    page: number | null,
    startOffset: number | null,
    endOffset: number | null,
    text: string | null,
    annotation: string | null,
    annotationExtra: string | null,
    date: string | null,
    uri: string | null,
    createdAt: string,
    updatedAt: string,
    owner: string | null,
  } | null,
};

export type DeleteHighlightMutationVariables = {
  input: DeleteHighlightInput,
  condition?: ModelHighlightConditionInput | null,
};

export type DeleteHighlightMutation = {
  deleteHighlight:  {
    __typename: "Highlight",
    id: string,
    type: HighlightType,
    deviceType: DeviceType,
    title: string,
    isbn: string | null,
    author: string | null,
    page: number | null,
    startOffset: number | null,
    endOffset: number | null,
    text: string | null,
    annotation: string | null,
    annotationExtra: string | null,
    date: string | null,
    uri: string | null,
    createdAt: string,
    updatedAt: string,
    owner: string | null,
  } | null,
};

export type CreateDocumentMutationVariables = {
  input: CreateDocumentInput,
  condition?: ModelDocumentConditionInput | null,
};

export type CreateDocumentMutation = {
  createDocument:  {
    __typename: "Document",
    id: string,
    library: string,
    name: string,
    fileName: string,
    type: DeviceType | null,
    createdAt: string,
    updatedAt: string,
    owner: string | null,
  } | null,
};

export type UpdateDocumentMutationVariables = {
  input: UpdateDocumentInput,
  condition?: ModelDocumentConditionInput | null,
};

export type UpdateDocumentMutation = {
  updateDocument:  {
    __typename: "Document",
    id: string,
    library: string,
    name: string,
    fileName: string,
    type: DeviceType | null,
    createdAt: string,
    updatedAt: string,
    owner: string | null,
  } | null,
};

export type DeleteDocumentMutationVariables = {
  input: DeleteDocumentInput,
  condition?: ModelDocumentConditionInput | null,
};

export type DeleteDocumentMutation = {
  deleteDocument:  {
    __typename: "Document",
    id: string,
    library: string,
    name: string,
    fileName: string,
    type: DeviceType | null,
    createdAt: string,
    updatedAt: string,
    owner: string | null,
  } | null,
};

export type GetHighlightQueryVariables = {
  id: string,
};

export type GetHighlightQuery = {
  getHighlight:  {
    __typename: "Highlight",
    id: string,
    type: HighlightType,
    deviceType: DeviceType,
    title: string,
    isbn: string | null,
    author: string | null,
    page: number | null,
    startOffset: number | null,
    endOffset: number | null,
    text: string | null,
    annotation: string | null,
    annotationExtra: string | null,
    date: string | null,
    uri: string | null,
    createdAt: string,
    updatedAt: string,
    owner: string | null,
  } | null,
};

export type ListHighlightsQueryVariables = {
  filter?: ModelHighlightFilterInput | null,
  limit?: number | null,
  nextToken?: string | null,
};

export type ListHighlightsQuery = {
  listHighlights:  {
    __typename: "ModelHighlightConnection",
    items:  Array< {
      __typename: "Highlight",
      id: string,
      type: HighlightType,
      deviceType: DeviceType,
      title: string,
      isbn: string | null,
      author: string | null,
      page: number | null,
      startOffset: number | null,
      endOffset: number | null,
      text: string | null,
      annotation: string | null,
      annotationExtra: string | null,
      date: string | null,
      uri: string | null,
      createdAt: string,
      updatedAt: string,
      owner: string | null,
    } | null > | null,
    nextToken: string | null,
  } | null,
};

export type GetDocumentQueryVariables = {
  id: string,
};

export type GetDocumentQuery = {
  getDocument:  {
    __typename: "Document",
    id: string,
    library: string,
    name: string,
    fileName: string,
    type: DeviceType | null,
    createdAt: string,
    updatedAt: string,
    owner: string | null,
  } | null,
};

export type ListDocumentsQueryVariables = {
  filter?: ModelDocumentFilterInput | null,
  limit?: number | null,
  nextToken?: string | null,
};

export type ListDocumentsQuery = {
  listDocuments:  {
    __typename: "ModelDocumentConnection",
    items:  Array< {
      __typename: "Document",
      id: string,
      library: string,
      name: string,
      fileName: string,
      type: DeviceType | null,
      createdAt: string,
      updatedAt: string,
      owner: string | null,
    } | null > | null,
    nextToken: string | null,
  } | null,
};

export type OnCreateHighlightSubscriptionVariables = {
  owner: string,
};

export type OnCreateHighlightSubscription = {
  onCreateHighlight:  {
    __typename: "Highlight",
    id: string,
    type: HighlightType,
    deviceType: DeviceType,
    title: string,
    isbn: string | null,
    author: string | null,
    page: number | null,
    startOffset: number | null,
    endOffset: number | null,
    text: string | null,
    annotation: string | null,
    annotationExtra: string | null,
    date: string | null,
    uri: string | null,
    createdAt: string,
    updatedAt: string,
    owner: string | null,
  } | null,
};

export type OnUpdateHighlightSubscriptionVariables = {
  owner: string,
};

export type OnUpdateHighlightSubscription = {
  onUpdateHighlight:  {
    __typename: "Highlight",
    id: string,
    type: HighlightType,
    deviceType: DeviceType,
    title: string,
    isbn: string | null,
    author: string | null,
    page: number | null,
    startOffset: number | null,
    endOffset: number | null,
    text: string | null,
    annotation: string | null,
    annotationExtra: string | null,
    date: string | null,
    uri: string | null,
    createdAt: string,
    updatedAt: string,
    owner: string | null,
  } | null,
};

export type OnDeleteHighlightSubscriptionVariables = {
  owner: string,
};

export type OnDeleteHighlightSubscription = {
  onDeleteHighlight:  {
    __typename: "Highlight",
    id: string,
    type: HighlightType,
    deviceType: DeviceType,
    title: string,
    isbn: string | null,
    author: string | null,
    page: number | null,
    startOffset: number | null,
    endOffset: number | null,
    text: string | null,
    annotation: string | null,
    annotationExtra: string | null,
    date: string | null,
    uri: string | null,
    createdAt: string,
    updatedAt: string,
    owner: string | null,
  } | null,
};

export type OnCreateDocumentSubscriptionVariables = {
  owner: string,
};

export type OnCreateDocumentSubscription = {
  onCreateDocument:  {
    __typename: "Document",
    id: string,
    library: string,
    name: string,
    fileName: string,
    type: DeviceType | null,
    createdAt: string,
    updatedAt: string,
    owner: string | null,
  } | null,
};

export type OnUpdateDocumentSubscriptionVariables = {
  owner: string,
};

export type OnUpdateDocumentSubscription = {
  onUpdateDocument:  {
    __typename: "Document",
    id: string,
    library: string,
    name: string,
    fileName: string,
    type: DeviceType | null,
    createdAt: string,
    updatedAt: string,
    owner: string | null,
  } | null,
};

export type OnDeleteDocumentSubscriptionVariables = {
  owner: string,
};

export type OnDeleteDocumentSubscription = {
  onDeleteDocument:  {
    __typename: "Document",
    id: string,
    library: string,
    name: string,
    fileName: string,
    type: DeviceType | null,
    createdAt: string,
    updatedAt: string,
    owner: string | null,
  } | null,
};
