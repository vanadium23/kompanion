# 4. Design OPDS

Date: 2024-09-02

## Status

Accepted

## Context

In order for KOReader to be able to download books from the application, it is necessary to implement the Open Publication Distribution System (OPDS). The official website of the specification is here - https://specs.opds.io/. The current format consists of Atom + XML over HTTP.

## Decision

1. It is necessary to implement the specification - https://specs.opds.io/opds-1.2
2. OPDS is an alternative to the web API for usecase.Books
3. OPDS is only concerned with the formatting of books and access to this API
4. OPDS will be located at the URL `/opds`
5. OPDS will be secured with basic authentication

Given all of this, it seems that OPDS is simply a controller.

### Design URLs

The general approach is that OPDS does not heavily regulate the operation of URLs. Here, it resembles RESTful in the canonical understanding, where, in addition to the list of entities, links to adjacent pages must also be provided. The design was taken from Calibre, and it represents a tree structure:
- /opds - list of shelves sorted/filtered
    - /opds/<:shelf>/ - specific sorting/filtering with links forward and backward
        - the download link leads to v1/downloadBook (connectivity?)
    - /opds/search/<:search>/ - search for books by a string

## Consequences

Risks:
- There may be logic in OPDS (for example, sorting selection) that should not be in the controller.
