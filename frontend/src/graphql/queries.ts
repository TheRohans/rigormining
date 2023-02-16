/* tslint:disable */
/* eslint-disable */
// this is an auto generated file. This will be overwritten

export const getHighlight = /* GraphQL */ `
  query GetHighlight($id: ID!) {
    getHighlight(id: $id) {
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
export const listHighlights = /* GraphQL */ `
  query ListHighlights(
    $filter: ModelHighlightFilterInput
    $limit: Int
    $nextToken: String
  ) {
    listHighlights(filter: $filter, limit: $limit, nextToken: $nextToken) {
      items {
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
      nextToken
    }
  }
`;
export const getDocument = /* GraphQL */ `
  query GetDocument($id: ID!) {
    getDocument(id: $id) {
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
export const listDocuments = /* GraphQL */ `
  query ListDocuments(
    $filter: ModelDocumentFilterInput
    $limit: Int
    $nextToken: String
  ) {
    listDocuments(filter: $filter, limit: $limit, nextToken: $nextToken) {
      items {
        id
        library
        name
        fileName
        type
        createdAt
        updatedAt
        owner
      }
      nextToken
    }
  }
`;
