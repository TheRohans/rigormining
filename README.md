# Knotset

Your dumping ground for research

Possible [long term vision](https://miro.com/app/board/uXjVOzAdL8A=/)

```c
struct highlights
{
    // Where this came from
    char device_type[8];
    // Generated hash id
    char *id;
    // start offset in the ebook
    int start_offset;
    // end offset in the ebook
    int end_offset;
    // page of highlight (if it exists)
    int page;
    // Actual text of the highlight
    char *text;
    // text annotation
    char *annotation;
    // ...
    char *extra_annotation_data;
    // 2021-05-22T04:32:40.769
    char *date_created;
    // Internal ID of this content (if it exists)
    // file:///mnt/onboard/Ekert, Artur & Hosgood, Tim/Lectures on Quantum Information Science - Artur
    // Ekert & Tim Hosgood.epub#(5)EPUB/text/ch003.xhtml#some-mathematical-preliminaries
    char *content_id;
    // Title of the work (book)
    char *title;
    // Book ISBN (if it exists)
    char *ISBN;
    // Author
    char *attribution;
    // Note or Highlight
    char highlight_type[12];
};
```

```javascript
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
```


## Getting Started

As I am sure you'd guess, the backend and the frontend need to run at the same time. I just use different tabs in me shell.

### Backend

Start up the "backend". This currently just serves up a document listing and the documents to test out the PDF and EPub viewer.

```
cd backend
make install
make start
```

This serves documents on http://localhost:8000 and currently just out of the "rohan" public library - for example `http://localhost:8000/public/library/rohan/metadata.json`

### Frontend

```
cd frontend
make install
make start
```

It takes a minute, but if all goes well, you should be able to view the site by going to: `http://localhost:8080/`.

You can login with anything for credentials.

I've removed most of the original backend for this - it was using AWS amplify, graphql, and dynmodb for login and storage, but the amplify version of those things suck. So any reference to graphql don't work, but are there for reference.

The only thing that still works right now, is the PDF and epub document viewing - you can see highlighting callback in the console.
