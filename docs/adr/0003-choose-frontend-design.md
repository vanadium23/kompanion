# 3. Choose Frontend Design

Date: 2024-08-19

## Status

Accepted

## Context

Most of the APIs were implemented during the transition to [go-clean-template](./0002-use-evrone-golang-template.md). Now we need to think about how to implement the frontend.
It is important to understand that my skills are not sufficient for developing my own design. Additionally, I would not like to spend time managing a bunch of `node_modules/` through `vite` or some bundler. In other words, the more straightforward, the better.

## Decision

The change that we're proposing or have agreed to implement:
- We will use vanilla HTML + JS
- From the perspective of the codebase, we will create a controller/web, where the response consists of templates
- The templates will be created using html/template to avoid pulling in dependencies

## Alternatives

### templ

templ is a React-like engine in Golang. The essence of its operation is as follows: you write files with the .templ extension, then code generation occurs, and then you plug it into your code like a regular package.

Pros:
- The frontend is written in Golang

Cons:
- Additional dependency
- Code generation stage
- CSS/layout integration is more difficult than in React

### htmx

htmx is a new trendy hotwire or turbopages. You connect one script, and then you add tags to vanilla HTML to fetch partial HTML. The only thing that is unclear is how to handle authorization in this case.

Pros:
- Only one JS file

Cons:
- Template logic spreads between the backend and frontend (re-check)
- JS logic will still be present, or there will be ugly blocks in HTML attributes

### Vue in One File

Vue 3 is a full-fledged JS framework with builds and everything else. However, Julia Evans created a build to have it all on one page: https://jvns.ca/blog/2023/02/16/writing-javascript-without-a-build-system/

Pros:
- A full-fledged JS framework without a build

Cons:
- The code will be in a monolithic JS file
- You can't easily integrate a router


## Consequences

- (+) no logic client side
- (-) upload book file will be a very long reload for page
