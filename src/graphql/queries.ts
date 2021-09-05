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
      text
      annotation
      annotationExtra
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
        text
        annotation
        annotationExtra
        createdAt
        updatedAt
        owner
      }
      nextToken
    }
  }
`;
