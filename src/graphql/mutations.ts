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
      uri
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
      uri
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
      uri
      createdAt
      updatedAt
      owner
    }
  }
`;
export const createDocument = /* GraphQL */ `
  mutation CreateDocument(
    $input: CreateDocumentInput!
    $condition: ModelDocumentConditionInput
  ) {
    createDocument(input: $input, condition: $condition) {
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
export const updateDocument = /* GraphQL */ `
  mutation UpdateDocument(
    $input: UpdateDocumentInput!
    $condition: ModelDocumentConditionInput
  ) {
    updateDocument(input: $input, condition: $condition) {
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
export const deleteDocument = /* GraphQL */ `
  mutation DeleteDocument(
    $input: DeleteDocumentInput!
    $condition: ModelDocumentConditionInput
  ) {
    deleteDocument(input: $input, condition: $condition) {
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
