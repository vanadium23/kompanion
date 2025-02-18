# 7. Choose CSS Framework

Date: 2024-10-01

## Status

Accepted

## Context

The KOmpanion service has a frontend. Its main responsibility is to manage the library and nothing more. This means that no complex components are expected; moreover, as decided in [ADR 0003](./0003-choose-frontend-design.md), no JavaScript is needed. Additionally, the entire setup should be bundled into a single binary file. Therefore, to keep it lightweight, we need to:
1. Have a CSS framework in a single file, without JavaScript.
2. Ensure it does not have a dozen helper classes and works on top of standard tags.

To make a choice, we need to determine the "components" that I will later require:
1. Authorization form
2. Table for books
3. Book images
4. Navbar
5. Footer

## Decision

Ultimately, a solution was found in the form of [classless CSS](https://github.com/dbohdan/classless-css?tab=readme-ov-file). The specific solution chosen is water.css, as it is the most minimalist in appearance.

Below is a comparison of all the solutions reviewed.

### [Basis.css](https://vladocar.github.io/Basic.css/).

Pros:
- Has a grid on `section`
- Dark theme

Cons:
- No header & footer

### [MVP.css](https://andybrewer.github.io/mvp/mvp.html)

Pros:
- Nice forms
- Has a quickstart that will fit

Cons:
- Too Bootstrap-like, and the name suggests it

### [Water.css](https://watercss.kognise.dev/)

Pros:
- 3 KB, and has dark/light themes
- Includes forms

Cons:
- No grid

### [Monospace Web](https://owickstrom.github.io/the-monospace-web/)

Pros:
- Has a grid
- Small size

Cons:
- Default font is JetBrains Mono, which needs to be downloaded

## Consequences

- (+) You can simply write semantic HTML
- (-) Customizing elements without styles may cause issues somewhere
