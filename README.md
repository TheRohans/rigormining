# Knotset

Your dumping ground for research

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

I've removed most of the original backend for this - it was using AWS amplify, graphql, and dynmodb for login and storage, but those amplify version of those things suck. So any reference you see to graphql doesn't work, but I left it in to see how it was working.
