# 2. Use evrone/golang-template

Date: 2024-08-19

## Status

Accepted

## Context

I don't have much experience with Golang projects, so I want to use a more structured implementation.
From the template, I need a solution for the following tasks:
- Organization of the codebase
- Solutions for integration testing
- Use of migrations for the database

## Decision

The change that we're proposing or have agreed to implement:
- Move kompanion to the template at https://gitea.chrnv.ru/vanadium23/kompanion

## Alternatives

### Start from scratch

There was an attempt, not necessarily a bad one. The source code can be viewed in the old-master branch.
I decided to abandon it because I couldn't understand:
- How to implement integration tests
- What to do with migrations

### Golang standards 

https://github.com/golang-standards/project-layout

It only describes 

## Consequences

From the template, I will have to remove everything that I do not need or that does not fit:
1. RabbitMQ and RPC on top of it
2. Probably Gin, as I want to use `net/http`
3. Reassess the list of dependencies from the template
