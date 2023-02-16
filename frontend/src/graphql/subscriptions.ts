/* tslint:disable */
/* eslint-disable */
// this is an auto generated file. This will be overwritten

export const onCreateHighlight = /* GraphQL */ `
  subscription OnCreateHighlight($owner: String!) {
    onCreateHighlight(owner: $owner) {
      id
      type
      deviceType
      title
      isbn
      author
      page
      startOffset
      endOffset
      text
      annotation
      annotationExtra
      date
      uri
      createdAt
      updatedAt
      owner
    }
  }
`;
export const onUpdateHighlight = /* GraphQL */ `
  subscription OnUpdateHighlight($owner: String!) {
    onUpdateHighlight(owner: $owner) {
      id
      type
      deviceType
      title
      isbn
      author
      page
      startOffset
      endOffset
      text
      annotation
      annotationExtra
      date
      uri
      createdAt
      updatedAt
      owner
    }
  }
`;
export const onDeleteHighlight = /* GraphQL */ `
  subscription OnDeleteHighlight($owner: String!) {
    onDeleteHighlight(owner: $owner) {
      id
      type
      deviceType
      title
      isbn
      author
      page
      startOffset
      endOffset
      text
      annotation
      annotationExtra
      date
      uri
      createdAt
      updatedAt
      owner
    }
  }
`;
export const onCreateDocument = /* GraphQL */ `
  subscription OnCreateDocument($owner: String!) {
    onCreateDocument(owner: $owner) {
      id
      library
      name
      fileName
      type
      createdAt
      updatedAt
      owner
    }
  }
`;
export const onUpdateDocument = /* GraphQL */ `
  subscription OnUpdateDocument($owner: String!) {
    onUpdateDocument(owner: $owner) {
      id
      library
      name
      fileName
      type
      createdAt
      updatedAt
      owner
    }
  }
`;
export const onDeleteDocument = /* GraphQL */ `
  subscription OnDeleteDocument($owner: String!) {
    onDeleteDocument(owner: $owner) {
      id
      library
      name
      fileName
      type
      createdAt
      updatedAt
      owner
    }
  }
`;
