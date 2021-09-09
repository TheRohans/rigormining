/* tslint:disable */
/* eslint-disable */
// this is an auto generated file. This will be overwritten

export const createHighlight = /* GraphQL */ `
  mutation CreateHighlight(
    $input: CreateHighlightInput!
    $condition: ModelHighlightConditionInput
  ) {
    createHighlight(input: $input, condition: $condition) {
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
      createdAt
      updatedAt
      owner
    }
  }
`;
export const updateHighlight = /* GraphQL */ `
  mutation UpdateHighlight(
    $input: UpdateHighlightInput!
    $condition: ModelHighlightConditionInput
  ) {
    updateHighlight(input: $input, condition: $condition) {
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
      createdAt
      updatedAt
      owner
    }
  }
`;
export const deleteHighlight = /* GraphQL */ `
  mutation DeleteHighlight(
    $input: DeleteHighlightInput!
    $condition: ModelHighlightConditionInput
  ) {
    deleteHighlight(input: $input, condition: $condition) {
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
      createdAt
      updatedAt
      owner
    }
  }
`;
